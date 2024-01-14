package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func Commit(imageName string) error {
	mntURL := "/root/merged"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Printf("commit ImageTar:%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Error("tar folder %s error %v", mntURL, err)
	}
	return nil
}
