// Clockfs implements a file system with the current time in a file.
// It was written to demonstrate kernel cache invalidation.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"



	"test.io/fuse/k8s-fuse/pkg"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"

)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

// func run(mountpoint string) error {
// 	c, err := fuse.Mount(
// 		mountpoint,
// 		fuse.FSName("clock"),
// 		fuse.Subtype("clockfsfs"),
// 		fuse.LocalVolume(),
// 		fuse.VolumeName("Clock filesystem"),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer c.Close()
//
// 	if p := c.Protocol(); !p.HasInvalidate() {
// 		return fmt.Errorf("kernel FUSE support is too old to have invalidations: version %v", p)
// 	}
//
// 	srv := fs.New(c, nil)
// 	filesys := &FS{
// 		// We pre-create the clock node so that it's always the same
// 		// object returned from all the Lookups. You could carefully
// 		// track its lifetime between Lookup&Forget, and have the
// 		// ticking & invalidation happen only when active, but let's
// 		// keep this example simple.
// 		clockFile: &File{
// 			fuse: srv,
// 		},
// 	}
// 	filesys.clockFile.tick()
// 	// This goroutine never exits. That's fine for this example.
// 	go filesys.clockFile.update()
// 	if err := srv.Serve(filesys); err != nil {
// 		return err
// 	}
//
// 	// Check if the mount process has an error to report.
// 	<-c.Ready
// 	if err := c.MountError; err != nil {
// 		return err
// 	}
// 	return nil
// }

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}

	currentdir,_:=os.Getwd()
    	mountpoint := flag.Arg(0)
    	if strings.HasPrefix(mountpoint,"/"){
    	pkg.MountDir=mountpoint
    	}else{
    	pkg.MountDir,_=filepath.Abs(filepath.Join(currentdir,mountpoint))
    	}



	c, err := fuse.Mount(
    		mountpoint,
    		fuse.FSName("kubernetes"),
    		fuse.Subtype("k8sfs"),
    		fuse.LocalVolume(),
    		fuse.VolumeName("k8s"),
    	)
    	if err != nil {
    		log.Fatal(err)
    	}
    	defer c.Close()
        pkg.Init()

    	err = fs.Serve(c, &pkg.FS{})
    	if err != nil {
    		log.Fatal(err)
    	}

    	// check if the mount process has an error to report
    	<-c.Ready
    	if err := c.MountError; err != nil {
    		log.Fatal(err)
    	}
}
