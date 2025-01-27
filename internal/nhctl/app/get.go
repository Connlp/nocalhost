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

package app

import (
	"nocalhost/internal/nhctl/nocalhost"
	"nocalhost/internal/nhctl/profile"
	"nocalhost/pkg/nhctl/log"
	"os"
	"path/filepath"
)

func (a *Application) GetDependencies() []*SvcDependency {
	result := make([]*SvcDependency, 0)

	svcConfigs := a.appMeta.Config.ApplicationConfig.ServiceConfigs
	if len(svcConfigs) == 0 {
		return nil
	}

	for _, svcConfig := range svcConfigs {
		if svcConfig.DependLabelSelector == nil {
			continue
		}
		if svcConfig.DependLabelSelector.Pods == nil && svcConfig.DependLabelSelector.Jobs == nil {
			continue
		}
		svcDep := &SvcDependency{
			Name: svcConfig.Name,
			Type: svcConfig.Type,
			Jobs: svcConfig.DependLabelSelector.Jobs,
			Pods: svcConfig.DependLabelSelector.Pods,
		}
		result = append(result, svcDep)
	}
	return result
}

// Get local path of resource dirs
// If resource path undefined, use git url
func (a *Application) GetResourceDir(tmpDir string) []string {
	appProfile, _ := a.GetProfile()
	var resourcePath []string
	if len(appProfile.ResourcePath) != 0 {
		for _, path := range appProfile.ResourcePath {
			fullPath := filepath.Join(tmpDir, path)
			resourcePath = append(resourcePath, fullPath)
		}
		return resourcePath
	}
	return []string{tmpDir}
}

func (a *Application) getIgnoredPath() []string {
	appProfile, _ := a.GetProfile()
	results := make([]string, 0)
	for _, path := range appProfile.IgnoredPath {
		results = append(results, filepath.Join(a.ResourceTmpDir, path))
	}
	return results
}

func (a *Application) GetDefaultWorkDir(svcName, container string) string {
	svcProfile, _ := a.GetSvcProfile(svcName)
	if svcProfile != nil && svcProfile.GetContainerDevConfigOrDefault(container).WorkDir != "" {
		return svcProfile.GetContainerDevConfigOrDefault(container).WorkDir
	}
	return profile.DefaultWorkDir
}

func (a *Application) GetPersistentVolumeDirs(svcName, container string) []*profile.PersistentVolumeDir {
	svcProfile, _ := a.GetSvcProfile(svcName)
	if svcProfile != nil {
		return svcProfile.GetContainerDevConfigOrDefault(container).PersistentVolumeDirs
	}
	return nil
}

//func (a *Application) GetDefaultSideCarImage(svcName string) string {
//	return DefaultSideCarImage
//}

func (a *Application) GetDefaultDevImage(svcName string, container string) string {
	svcProfile, _ := a.GetSvcProfile(svcName)
	if svcProfile != nil && svcProfile.GetContainerDevConfigOrDefault(container).Image != "" {
		return svcProfile.GetContainerDevConfigOrDefault(container).Image
	}
	return profile.DefaultDevImage
}

func (a *Application) GetDefaultDevPort(svcName string, container string) []string {
	svcProfile, _ := a.GetSvcProfile(svcName)
	if svcProfile != nil && len(svcProfile.GetContainerDevConfigOrDefault(container).PortForward) > 0 {
		return svcProfile.GetContainerDevConfigOrDefault(container).PortForward
	}
	return []string{}
}

func (a *Application) GetApplicationSyncDir(deployment string) string {
	dirPath := filepath.Join(a.GetHomeDir(), nocalhost.DefaultBinSyncThingDirName, deployment)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0700)
		if err != nil {
			log.Fatalf("fail to create syncthing directory: %s", dirPath)
		}
	}
	return dirPath
}
