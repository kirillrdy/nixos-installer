package main

import (
	"log"
	"os"
	"os/exec"
)

func crash(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func sh(cmdName string, args ...string) {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	crash(err)
}

func main() {
	targetDevice = "/dev/sdb"
  rootPartition = targetDevice+ "1"
  swapPartition = targetDevice+ "2"
  bootPartition = targetDevice+ "2"
	sh("ls")
	sh("parted", targetDevice, "--", "mklabel", "gpt")
	sh("parted", targetDevice, "--", "mkpart", "primary", "512MiB", "-8GiB")
	sh("parted", targetDevice, "--", "mkpart", "primary", "linux-swap", "-8GiB", "100%")
	sh("parted", targetDevice, "--", "mkpart", "ESP", "fat32", "1MiB", "512MiB")
	sh("parted", targetDevice, "--", "set", "3", "esp", "on")
	sh("mkfs.ext4", rootPartition)
	sh("mkswap", swapPartition)
	sh("mkfs.fat", "-F", "32", "-n", "boot", bootPartition)
	sh("mount", rootPartition ,"/mnt")
	sh("mkdir", "-p", "/mnt/boot")
	sh("mount", bootPartition ,"/mnt/boot")
	sh("swapon" swapPartition)
	sh("nixos-generate-config", "--root", "/mnt")
	sh("nixos-install")

}
