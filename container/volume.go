package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

/*
容器的文件系统相关操作
*/

// NewWorkSpace Create an Overlay2 filesystem as container root workspace
func NewWorkSpace(rootURL string, mntURL string, volume string) {
	createLower(rootURL)
	createDirs(rootURL)
	mountOverlayFS(rootURL, mntURL)
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volume[1] != "" {
			mountVolume(rootURL, mntURL, volumeURLs)
			log.Infof("[NewWorkSpace] volumeURL:%q", volumeURLs)
		} else {
			log.Infof("[NewWorkSpace] volume parameter input is not correct.")
		}

	}
}

// createLower 将busybox作为overlayfs的lower层
func createLower(rootURL string) {
	// 把busybox作为overlayfs中的lower层
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	// 检查是否已经存在busybox文件夹
	exist, err := pathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	// 不存在则创建目录并将busybox.tar解压到busybox文件夹中
	if !exist {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", busyboxURL, err)
		}
	}
}

// createDirs 创建overlayfs需要的的upper、worker目录
func createDirs(rootURL string) {
	upperURL := rootURL + "upper/"
	if err := os.Mkdir(upperURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", upperURL, err)
	}
	workURL := rootURL + "work/"
	if err := os.Mkdir(workURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", workURL, err)
	}
}

// mountOverlayFS 挂载overlayfs
func mountOverlayFS(rootURL string, mntURL string) {
	// mount -t overlay overlay -o lowerdir=lower1:lower2:lower3,upperdir=upper,workdir=work merged
	// 创建对应的挂载目录
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", mntURL, err)
	}
	// 拼接参数
	// e.g. lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/merged
	dirs := "lowerdir=" + rootURL + "busybox" + ",upperdir=" + rootURL + "upper" + ",workdir=" + rootURL + "work"
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	// 如果指定了volume则需要先umount volume
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			umountVolume(mntURL, volumeURLs)
		}
	}
	umountOverlayFS(mntURL)
	deleteDirs(rootURL)
}

func umountOverlayFS(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func deleteDirs(rootURL string) {
	writeURL := rootURL + "upper/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
	workURL := rootURL + "work"
	if err := os.RemoveAll(workURL); err != nil {
		log.Errorf("Remove dir %s error %v", workURL, err)
	}
}

// pathExists returns whether a path exists.
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// volumeUrlExtract 通过冒号分割解析volume目录，比如 -v /tmp:/tmp
func volumeUrlExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

func mountVolume(rootURL string, mntURL string, volumeURLs []string) {
	// 第0个元素为宿主机目录
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		log.Infof("[mountVolume] mkdir parent dir %s error. %v", parentUrl, err)
	}
	// 第1个元素为容器目录
	containerUrl := volumeURLs[1]
	// 拼接并创建对应的容器目录
	containerVolumeURL := mntURL + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("[mountVolume] mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	// 通过bind mount 将宿主机目录挂载到容器目录
	// mount -o bind /hostURL /containerURL
	cmd := exec.Command("mount", "-o", "bind", parentUrl, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("[mountVolume] mount volume failed. %v", err)
	}
}

func umountVolume(mntURL string, volumeURLs []string) {
	// 卸载容器里volume挂载点的文件系统
	containerURL := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount volume faied. %v", err)
	}
}
