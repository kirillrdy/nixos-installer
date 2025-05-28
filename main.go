package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

const zfs = "zfs"
const ext4 = "ext4"

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func execute(cmdName string, args ...string) {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	check(cmd.Run())
}

func main() {

	compression := flag.Bool("compress", true, "use compression on ZFS pool")
	encryption := flag.Bool("encrypt", false, "use encryption on ZFS pool")
	swapSize := flag.Int("swap", 32, "swap size in GiB")

	rootFileSystem := flag.String("fs", zfs, "filesystem to use on root, currently ext4 and zfs")
	targetDevice := flag.String("device", "", "Device to use ")
	flag.Parse()

	if *targetDevice == "" {
		fmt.Print("Please provide device name via -device flag")
		os.Exit(1)
	}

	rootPartition := *targetDevice + "1"
	swapPartition := *targetDevice + "2"
	bootPartition := *targetDevice + "3"

	if strings.HasPrefix(*targetDevice, "/dev/nvme") {
		rootPartition = *targetDevice + "p1"
		swapPartition = *targetDevice + "p2"
		bootPartition = *targetDevice + "p3"
	}

	swapPartitionSize := fmt.Sprint("-", *swapSize, "GiB")

	execute("parted", *targetDevice, "--", "mklabel", "gpt")
	execute("parted", *targetDevice, "--", "mkpart", "primary", "512MiB", swapPartitionSize)
	execute("parted", *targetDevice, "--", "mkpart", "primary", "linux-swap", swapPartitionSize, "100%")
	execute("parted", *targetDevice, "--", "mkpart", "ESP", "fat32", "1MiB", "512MiB")
	execute("parted", *targetDevice, "--", "set", "3", "esp", "on")
	if *rootFileSystem == ext4 {
		execute("mkfs.ext4", rootPartition)
		execute("mount", rootPartition, "/mnt")
	} else if *rootFileSystem == zfs {
		zfsPoolName := "zroot"
		nixosZfsDataset := path.Join(zfsPoolName, "root")

		createArgs := []string{
			"create", "-f", "-O", "mountpoint=none", "-O", "atime=off",
		}
		if *compression {
			createArgs = append(createArgs, "-O", "compression=zstd")
		}

		if *encryption {
			createArgs = append(createArgs, "-O", "encryption=aes-256-gcm")
			createArgs = append(createArgs, "-O", "keyformat=passphrase")
		}
		createArgs = append(createArgs, "-O", "xattr=sa", "-O", "acltype=posixacl", "-o", "autotrim=on", "-o", "ashift=12", "-R", "/mnt", zfsPoolName, rootPartition)

		execute("zpool", createArgs...)
		execute("zfs", "create", "-o", "mountpoint=legacy", nixosZfsDataset)
		execute("mount", "-t", "zfs", nixosZfsDataset, "/mnt")
	}

	execute("mkswap", swapPartition)
	execute("mkfs.fat", "-F", "32", "-n", "boot", bootPartition)

	execute("mkdir", "-p", "/mnt/boot")
	execute("mount", bootPartition, "/mnt/boot")
	execute("swapon", swapPartition)
	execute("nixos-generate-config", "--root", "/mnt")

	configFilePath := "/mnt/etc/nixos/configuration.nix"
	content, err := os.ReadFile(configFilePath)
	check(err)
	regex := regexp.MustCompile("\n{\n")
	newConfig := regex.ReplaceAllString(string(content), "\n{\n  networking.hostId = \"00000000\";\n")
	//TODO correct permissions
	check(os.WriteFile(configFilePath, []byte(newConfig), os.ModePerm))

	execute("nixos-install")

}
