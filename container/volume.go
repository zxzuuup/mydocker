package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

/*
容器的文件系统相关操作
*/

// NewWorkSpace Create an Overlay2 filesystem as container root workspace
func NewWorkSpace(rootURL string, mntURL string) {
	createLower(rootURL)
	createDirs(rootURL)
	mountOverlayFS(rootURL, mntURL)
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
func DeleteWorkSpace(rootURL string, mntURL string) {
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
