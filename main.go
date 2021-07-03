package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
)

const Zfs = "zfs"
const Ext4 = "ext4"

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

	isNvme := flag.Bool("nvme", false, "is target device NVME")
	rootFileSystem := flag.String("fs", Zfs, "filesystem to use on root, currently ext4 and zfs")
	targetDevice := flag.String("device", "", "Device to use ")
	flag.Parse()

	rootPartition := *targetDevice + "1"
	swapPartition := *targetDevice + "2"
	bootPartition := *targetDevice + "3"

	if *isNvme {
		rootPartition = *targetDevice + "n1p1"
		swapPartition = *targetDevice + "n1p2"
		bootPartition = *targetDevice + "n1p3"
	}

	sh("parted", *targetDevice, "--", "mklabel", "gpt")
	sh("parted", *targetDevice, "--", "mkpart", "primary", "512MiB", "-8GiB")
	sh("parted", *targetDevice, "--", "mkpart", "primary", "linux-swap", "-8GiB", "100%")
	sh("parted", *targetDevice, "--", "mkpart", "ESP", "fat32", "1MiB", "512MiB")
	sh("parted", *targetDevice, "--", "set", "3", "esp", "on")
	if *rootFileSystem == Ext4 {
		sh("mkfs.ext4", rootPartition)
		sh("mount", rootPartition, "/mnt")
	} else if *rootFileSystem == Zfs {
		zfsPoolName := "zroot"
		nixosZfsDataset := path.Join(zfsPoolName, "root")
		sh("zpool", "create", "-O", "mountpoint=none", "-O", "atime=off", "-O", "compression=zstd", "-O", "xattr=sa", "-O", "acltype=posixacl", "-o", "ashift=12", "-R", "/mnt", zfsPoolName, rootPartition)
		sh("zfs", "create", "-o", "mountpoint=legacy", nixosZfsDataset)
		sh("mount", "-t", "zfs", nixosZfsDataset, "/mnt")
	}

	sh("mkswap", swapPartition)
	sh("mkfs.fat", "-F", "32", "-n", "boot", bootPartition)

	sh("mkdir", "-p", "/mnt/boot")
	sh("mount", bootPartition, "/mnt/boot")
	sh("swapon", swapPartition)
	sh("nixos-generate-config", "--root", "/mnt")
	// TODO patch networkId
	sh("nixos-install")

}
