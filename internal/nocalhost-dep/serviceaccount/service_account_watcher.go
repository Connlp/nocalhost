/*
 * Tencent is pleased to support the open source community by making Nocalhost available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service_account

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"nocalhost/internal/nhctl/appmeta"
	"nocalhost/internal/nhctl/watcher"
	"nocalhost/pkg/nocalhost-api/pkg/clientgo"
	"nocalhost/pkg/nocalhost-api/pkg/setupcluster"
	"strings"
	"sync"
)

type ServiceAccountWatcher struct {
	clientset *kubernetes.Clientset

	cache *set /* serviceAccount */
	lock  sync.Mutex
	quit  chan bool

	watchController *watcher.Controller
}

func (saw *ServiceAccountWatcher) CreateOrUpdate(key string, obj interface{}) error {
	if sa, ok := obj.(*corev1.ServiceAccount); ok {
		return saw.join(sa)
	} else {
		errInfo := fmt.Sprintf("Fetching service account with key %s but could not cast to sa: %v", key, obj)
		glog.Error(errInfo)
		return fmt.Errorf(errInfo)
	}
}

func (saw *ServiceAccountWatcher) Delete(key string) error {
	appName, err := appmeta.GetApplicationName(key)
	if err != nil {
		return err
	}

	saw.left(appName)
	return nil
}

func (saw *ServiceAccountWatcher) WatcherInfo() string {
	return fmt.Sprintf("'ServiceAccount'")
}

func (saw *ServiceAccountWatcher) join(sa *corev1.ServiceAccount) error {
	for key := range sa.Labels {
		if key == clientgo.NocalhostLabel {
			isClusterAdmin, _ := saw.isClusterAdmin(sa)
			saw.cache.record(string(sa.UID), isClusterAdmin, sa.Name)
			glog.Infof(
				"ServiceAccountCache: refresh nocalhost sa in ns: %s, is cluster admin: %t", sa.Namespace,
				isClusterAdmin,
			)
			return nil
		}
	}
	return nil
}

func (saw *ServiceAccountWatcher) left(saName string) {
	if idx := strings.Index(saName, "/"); idx > 0 {
		if len(saName) > idx+1 {
			sa := saName[idx+1:]
			glog.Infof("ServiceAccountCache: remove nocalhost sa in ns: %s", saName[:idx])
			saw.cache.removeByServiceAccountName(sa)
		}
	}
}

func NewServiceAccountWatcher(clientset *kubernetes.Clientset) *ServiceAccountWatcher {
	return &ServiceAccountWatcher{
		clientset: clientset,
		cache:     newSet(),
		quit:      make(chan bool),
	}
}

type set struct {
	inner  map[string] /* UID */ bool              /* is cluster admin */
	helper map[string] /* serviceAccount */ string /* UID */
	lock   sync.Mutex
}

func newSet() *set {
	return &set{
		map[string] /* UID */ bool /* is cluster admin */ {},
		map[string] /* serviceAccount */ string /* UID */ {},
		sync.Mutex{},
	}
}

func (s *set) record(key string, isClusterAdmin bool, saName string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.inner[key] = isClusterAdmin
	s.helper[saName] = key
}

func (s *set) removeByServiceAccountName(saName string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if uid, ok := s.helper[saName]; ok {
		delete(s.inner, uid)
		delete(s.helper, saName)
	}
}

func (saw *ServiceAccountWatcher) isClusterAdmin(sa *corev1.ServiceAccount) (bool, error) {
	if len(sa.Secrets) == 0 {
		return false, nil
	}

	secret, err := saw.clientset.CoreV1().Secrets(sa.Namespace).Get(
		context.TODO(), sa.Secrets[0].Name, metav1.GetOptions{},
	)
	if err != nil {
		glog.Error(err)
		return false, err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Error(err)
		return false, err
	}

	KubeConfigYaml, err, _ := setupcluster.NewDevKubeConfigReader(
		secret, config.Host, sa.Namespace,
	).GetCA().GetToken().AssembleDevKubeConfig().ToYamlString()
	if err != nil {
		glog.Error(err)
		return false, err
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(KubeConfigYaml))
	if err != nil {
		glog.Error(err)
		return false, nil
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Error(err)
		return false, nil
	}

	arg := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: "*",
				Group:     "*",
				Verb:      "*",
				Name:      "*",
				Version:   "*",
				Resource:  "*",
			},
		},
	}

	response, err := cs.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), arg, metav1.CreateOptions{})
	if err != nil {
		glog.Error(err)
		return false, err
	}
	return response.Status.Allowed, nil
}

func (saw *ServiceAccountWatcher) IsClusterAdmin(uid string) *bool {
	admin, ok := saw.cache.inner[uid]
	if ok {
		return &admin
	} else {
		return nil
	}
}

func (saw *ServiceAccountWatcher) Quit() {
	saw.quit <- true
}

func (saw *ServiceAccountWatcher) Prepare() error {
	// create the service account watcher
	saWatcher := cache.NewListWatchFromClient(
		saw.clientset.CoreV1().RESTClient(), "serviceaccounts", "default", fields.Everything(),
	)

	controller := watcher.NewController(saw, saWatcher, &corev1.ServiceAccount{})

	if list, err := saw.clientset.CoreV1().ServiceAccounts("default").List(
		context.TODO(), metav1.ListOptions{},
	); err == nil {
		for _, item := range list.Items {
			for key := range item.Labels {
				if key == clientgo.NocalhostLabel {
					isClusterAdmin, _ := saw.isClusterAdmin(&item)
					saw.cache.record(string(item.UID), isClusterAdmin, item.Name)
					glog.Infof(
						"ServiceAccountCache: refresh nocalhost sa in ns: %s, is cluster admin: %t", item.Namespace,
						isClusterAdmin,
					)
				}
			}
		}
	}

	saw.watchController = controller
	return nil
}

// this method will block until error occur
func (saw *ServiceAccountWatcher) Watch() {
	stop := make(chan struct{})
	defer close(stop)
	go saw.watchController.Run(1, stop)
	<-saw.quit
}
