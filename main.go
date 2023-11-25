package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func deleteDir(dirPath string) {
	if dirPath == "" {
		return
	}
	err := os.RemoveAll(dirPath)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Directory deleted:", dirPath)
	}
}

func printStruct(data []struct{ word, value string }) {
	if len(data) == 0 {
		return
	}

	// Find the length of the longest word
	maxWordLen := len(data[0].word)
	for _, item := range data {
		if len(item.word) > maxWordLen {
			maxWordLen = len(item.word)
		}
	}

	// Print the data with aligned spacing
	for _, item := range data {
		fmt.Printf("%-*s %s\n", maxWordLen, item.word, item.value)
	}
}

func build(targetGoFile string, buildDir string, nameFormat string, version string, goos string, goarch string, doZip string) (bool, string) { // bool: status(true=success, false=error), string: dirPath for cancel
	currentDir, err := os.Getwd()
	if err != nil {
		// Handle the error
		fmt.Println(err)
		return false, ""
	}
	name := path.Base(currentDir)

	var buildDirName string
	if nameFormat == "1" {
		buildDirName = goos
	} else if nameFormat == "2" {
		switch goos {
		case "windows":
			buildDirName = "win"
		case "darwin":
			buildDirName = "mac"
		default:
			buildDirName = goos
		}
	} else {
		if version == "" {
			buildDirName = fmt.Sprintf("%s_%s_%s", name, goos, goarch)
		} else {
			buildDirName = fmt.Sprintf("%s_%s_%s_%s", name, version, goos, goarch)
		}
	}

	output := fmt.Sprintf("%s/%s", buildDir, buildDirName)
	fmt.Printf("Output: %s\n", output)

	extension := ""
	if goos == "windows" {
		extension = ".exe"
	}

	err = os.MkdirAll(output, 0755) // 0755 is the permission mode (rwxr-xr-x)
	if err != nil {
		fmt.Println("Error code1:", err)
		return false, buildDir
	}
	os.Setenv("GOOS", goos)
	os.Setenv("GOARCH", goarch)
	var cmd *exec.Cmd
	var outputPath string
	if strings.Contains(targetGoFile, ".go") {
		outputPath = fmt.Sprintf("%s/%s%s", output, name, extension)
		cmd = exec.Command("go", "build", "-o", outputPath, targetGoFile)
	} else {
		outputPath = fmt.Sprintf("%s/%s%s", output, targetGoFile, extension)
		cmd = exec.Command("go", "build", "-C", targetGoFile, "-o", outputPath)
	}
	fmt.Printf("\nexecutable output path is %s\n", outputPath)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error code2:", err)
		return false, buildDir
	}

	if doZip == "y" {
		zipFileName := fmt.Sprintf("%s.zip", output)

		// Create a ZIP file
		zipFile, err := os.Create(zipFileName)
		if err != nil {
			panic(err)
		}
		defer zipFile.Close()

		// Create a new ZIP writer
		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		walker := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Check if it's a directory; skip it if it is
			if info.IsDir() {
				return nil
			}

			// Open the file to be added to the ZIP archive
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Create a writer for the file inside the ZIP archive, excluding the directory itself
			zipPath, err := filepath.Rel(output, path)
			if err != nil {
				return err
			}

			zipEntry, err := zipWriter.Create(zipPath)
			if err != nil {
				return err
			}

			// Copy the file's content into the ZIP archive entry
			_, err = io.Copy(zipEntry, file)
			if err != nil {
				return err
			}

			return nil
		}

		// Recursively walk the directory and add files to the ZIP archive
		err = filepath.Walk(output, walker)
		if err != nil {
			panic(err)
		}
	}

	return true, ""
}

func moveFile(sourcePath, destinationDirectory string) error {
	// Join the destination directory with the file name from the source path
	destinationPath := filepath.Join(destinationDirectory, filepath.Base(sourcePath))

	// Rename the file to move it
	err := os.Rename(sourcePath, destinationPath)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	args := os.Args

	executableName := filepath.Base(args[0])

	if len(args) < 2 {
		fmt.Printf("Usage: %s <.go file> <goos:optional> <goarch:optional>\n", executableName)
		fmt.Printf("%s <.go file> for QUICKSTART\n", executableName)
		return
	}

	originalGOOS := runtime.GOOS
	originalGOARCH := runtime.GOARCH

	var targetGoFile string = args[1]
	var goos string = runtime.GOOS
	var goarch string = runtime.GOARCH
	var quickstart_all bool = false

	if len(args) > 2 {
		goos = args[2]
	} else {
		fmt.Println("[gocompile:QUICKSTART] Choose the OS among these OS presets for quick start")
		data := []struct {
			word, value string
		}{
			{"all", "Build compiles for windows, win32, macos and linux"},
			{"windows", ".EXE for Windows 64-bit"},
			{"win32", ".EXE for Windows 32-bit x86"},
			{"macos", "Executable file for ARM64 Darwin"},
			{"linux", "Executable file for ARM64 Linux"},
		}
		printStruct(data)
		fmt.Println()
		var quickstartAnswer string
		fmt.Print("OS: (press enter to skip QUICKSTART) ")
		fmt.Scanln(&quickstartAnswer)
		if quickstartAnswer == "windows" {
			goos = "windows"
			goarch = "amd64"
		} else if quickstartAnswer == "win32" {
			goos = "windows"
			goarch = "386"
		} else if quickstartAnswer == "macos" {
			goos = "darwin"
			goarch = "arm64"
		} else if quickstartAnswer == "linux" {
			goos = "linux"
			goarch = "arm64"
		} else if quickstartAnswer == "all" {
			quickstart_all = true
		} else if quickstartAnswer == "" {
			fmt.Printf(`
			aix
			android
			darwin
			dragonfly
			freebsd
			hurd
			illumos
			ios
			js
			linux
			nacl
			netbsd
			openbsd
			plan9
			solaris
			windows
			zos

			GOOS: (%s) `, originalGOOS)
			fmt.Scanln(&goos)
			fmt.Printf(`
			386
			amd64
			amd64p32
			arm
			arm64
			arm64be
			armbe
			loong64
			mips
			mips64
			mips64le
			mips64p32
			mips64p32le
			mipsle
			ppc
			ppc64
			ppc64le
			riscv
			riscv64
			s390
			s390x
			sparc
			sparc64
			wasm

			GOARCH: (%s) `, originalGOARCH)
			fmt.Scanln(&goarch)
		}
	}

	if len(args) > 3 {
		goos = args[3]
	}

	if len(args) > 4 {
		goarch = args[4]
	}

	var nameFormat string
	fmt.Print("choose directory name format\n0(default) = {name}_{version}_{goos}_{goarch}\n1 = {goos}\n2 = (windows=win, darwin=mac, others=goos)\n\n(optional): ")
	fmt.Scanln(&nameFormat)

	var version string
	if nameFormat == "" {
		fmt.Print("version (optional): ")
		fmt.Scanln(&version)
	}

	var buildDir string
	fmt.Print("build directory name: (build) ")
	fmt.Scanln(&buildDir)
	if buildDir == "" {
		buildDir = "build"
	}

	var doZip string
	fmt.Print("zip after build? (y/n): (n) ")
	fmt.Scanln(&doZip)
	if doZip == "" {
		doZip = "n"
	}

	if quickstart_all == true {
		goos = "windows"
		goarch = "amd64"
		status, dir := build(targetGoFile, buildDir, nameFormat, version, goos, goarch, doZip)
		if status == false {
			os.Setenv("GOOS", originalGOOS)
			os.Setenv("GOARCH", originalGOARCH)
			deleteDir(dir)
			return
		}
		goos = "windows"
		goarch = "386"
		status, dir = build(targetGoFile, buildDir, nameFormat, version, goos, goarch, doZip)
		if status == false {
			os.Setenv("GOOS", originalGOOS)
			os.Setenv("GOARCH", originalGOARCH)
			deleteDir(dir)
			return
		}
		goos = "darwin"
		goarch = "arm64"
		status, dir = build(targetGoFile, buildDir, nameFormat, version, goos, goarch, doZip)
		if status == false {
			os.Setenv("GOOS", originalGOOS)
			os.Setenv("GOARCH", originalGOARCH)
			deleteDir(dir)
			return
		}
		goos = "linux"
		goarch = "arm64"
		status, dir = build(targetGoFile, buildDir, nameFormat, version, goos, goarch, doZip)
		if status == false {
			os.Setenv("GOOS", originalGOOS)
			os.Setenv("GOARCH", originalGOARCH)
			deleteDir(dir)
			return
		}
	} else {
		status, dir := build(targetGoFile, buildDir, nameFormat, version, goos, goarch, doZip)
		if status == false {
			os.Setenv("GOOS", originalGOOS)
			os.Setenv("GOARCH", originalGOARCH)
			deleteDir(dir)
			return
		}
	}
	os.Setenv("GOOS", originalGOOS)
	os.Setenv("GOARCH", originalGOARCH)
	fmt.Println("gocompile: successfully building compiles!")
}
