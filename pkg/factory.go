package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"text/tabwriter"
	"time"

	"bazil.org/fuse/fs"
	//"gopkg.in/yaml.v3"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/labels"
)

var (
	MountDir string
	rootDir  = &Dir{
		inode: 1,
	}
	ipsMap  = map[string]*Dir{}
	procMap = map[string]*Dir{}
)

const padding = 3

type factory struct {
	kInfo
}

func (f *factory) FS() *FS {
	return &FS{}
}

type FS struct{}

func (fs *FS) Root() (fs.Node, error) {
	go func() {
		for {

			global_inode = 1
			nodesdir := newDir("nodes")
			podsdir := newDir("pods")
			ipsDir:=newDir("ip")
			procDir:=newDir("proc")

			var (
				totalcpus     int64
				totalmemory   int64
				requestcpus   int64
				requestmemory int64

				nsRcs = map[string]*resources{}
				ipsMap  = map[string]*Dir{}
				procMap = map[string]*Dir{}
			)

			nss, err := schema.client.CoreV1().Namespaces().List(context.Background(), meta_v1.ListOptions{})
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			nsmap := map[string]*Dir{}
			for _, ns := range nss.Items {
				nsmap[ns.Name] = newDir(ns.Name)
				nsRcs[ns.Name] = &resources{}
			}
			dump := schema.db.Dump()
			for k, lnode := range dump.Nodes {
				nodeDir := newDir(k)
				node := lnode.Clone()

				totalcpus += node.AllocatableResource().MilliCPU
				totalmemory += node.AllocatableResource().Memory
				requestcpus += node.RequestedResource().MilliCPU
				requestmemory += node.RequestedResource().Memory

				detailf := newFile("rc.detail", func() ([]byte, error) {
					var buf bytes.Buffer
					w := tabwriter.NewWriter(&buf, 0, 0, padding, ' ', 0)
					fmt.Fprintf(w, "allowedPodNumber\t%d\t\n", node.AllowedPodNumber())
					fmt.Fprintf(w, "allocatableMilliCPU\t%d\t\n", node.AllocatableResource().MilliCPU)
					fmt.Fprintf(w, "allocatableMemory\t%dM\t\n", node.AllocatableResource().Memory/1024/1024)
					fmt.Fprintf(w, "requestedMilliCPU\t%d\t\n", node.RequestedResource().MilliCPU)
					fmt.Fprintf(w, "requestedMemory\t%dM\t\n", node.RequestedResource().Memory/1024/1024)
					w.Flush()
					return buf.Bytes(), nil
				})
				nodeDir.Add(detailf)

				infofile := newFile("info", func() (i []byte, err error) {
					return json.MarshalIndent(node.Node(), " ", " ")
				})
				nodeDir.Add(infofile)

				for _, pod := range lnode.Pods() {
					podLink := newLink(pod.Name, filepath.Join(MountDir, fmt.Sprintf("/pods/%s/%s", pod.Namespace, pod.Name)))
					nodeDir.Add(podLink)

					nsdir, ok := nsmap[pod.Namespace]
					if ok {
						lpod := pod.DeepCopy()
						if _, ok := nsRcs[pod.Namespace]; ok {
							for _, container := range pod.Spec.Containers {
								nsRcs[pod.Namespace].cpu += container.Resources.Requests.Cpu().MilliValue()
								nsRcs[pod.Namespace].memory += container.Resources.Requests.Memory().Value()
							}

						}
						podDir := newDir(pod.Name)
						infoFile := newFile("info", func() (i []byte, err error) {

							return json.MarshalIndent(lpod, " ", " ")
						})
						execFile := newFile("exec", func() (i []byte, err error) {
							return []byte(fmt.Sprintf(`#! /bin/bash
kubectl exec -it %s bash -n %s		
`, lpod.Name, lpod.Namespace)), nil
						})
						if pod.Spec.NodeName != "" {
							nodeLink := newLink(pod.Spec.NodeName, filepath.Join(MountDir, "/nodes/"+pod.Spec.NodeName))
							podDir.Add(nodeLink)
						}

						podDir.Add(infoFile)
						podDir.Add(execFile)
						nsdir.Add(podDir)

						if len(pod.Status.PodIP)>0{
							ipd,ok:=ipsMap[pod.Status.PodIP]
							if !ok{
								ipd=newDir(pod.Status.PodIP)
								ipsMap[pod.Status.PodIP]=ipd
							}
							ipd.Add(podLink)
						}

						if pod.Labels!=nil{
							proc:=pod.Labels["proc"]
							if len(proc)>0{
								pdir,ok:=procMap[proc]
								if !ok{
									pdir=newDir(proc)
									procMap[proc]=pdir
								}
								pdir.Add(podLink)
							}
						}

					}
				}
				nodesdir.Add(nodeDir)
			}

			//			pods, _ := schema.podLister.List(labels.Everything())
			//
			//			for _, pod := range pods {
			//				nsdir, ok := nsmap[pod.Namespace]
			//				if ok {
			//					lpod := pod.DeepCopy()
			//					podDir := newDir(pod.Name)
			//					infoFile := newFile("info", func() (i []byte, err error) {
			//
			//						return json.MarshalIndent(lpod, " ", " ")
			//					})
			//					execFile := newFile("exec", func() (i []byte, err error) {
			//						return []byte(fmt.Sprintf(`#! /bin/bash
			//kubectl exec -it %s bash -n %s
			//`, lpod.Name, lpod.Namespace)), nil
			//					})
			//					if pod.Spec.NodeName != "" {
			//						nodeLink := newLink(pod.Spec.NodeName, filepath.Join(MountDir, "/nodes/"+pod.Spec.NodeName))
			//						podDir.Add(nodeLink)
			//					}
			//
			//					podDir.Add(infoFile)
			//					podDir.Add(execFile)
			//					nsdir.Add(podDir)
			//				}
			//			}
			//
			for k, nsdir := range nsmap {
				if rc, ok := nsRcs[k]; ok {
					nsdir.Add(newFile("rc.details", func() (i []byte, err error) {
						out := fmt.Sprintf("CPU[Requested]      %d\nMemory[Requested]   %dG\n", rc.cpu/1000, rc.memory/1024/2024/1000)
						return []byte(out), nil
					}))
				}
				podsdir.Add(nsdir)
			}

			resources_file := newFile("rc.detail", func() (i []byte, err error) {
				var buf bytes.Buffer
				w := tabwriter.NewWriter(&buf, 0, 0, padding, ' ', 0)
				fmt.Fprintf(w, "CPU[Total]\t%d\t\n", totalcpus/1000)
				fmt.Fprintf(w, "Memory[Total]\t%dG\t\n", totalmemory/1024/1024/1000)
				fmt.Fprintf(w, "CPU[Requested]\t%d\t\n", requestcpus/1000)
				fmt.Fprintf(w, "Memory[Requested]\t%dG\t\n", requestmemory/1024/1024/1000)

				w.Flush()
				return buf.Bytes(), nil
			})

			for _,dir:=range ipsMap{
				ipsDir.Add(dir)
			}

			for _,dir:=range procMap{
				procDir.Add(dir)
			}
			rootDir.children = nil
			rootDir.Add(ipsDir)
			rootDir.Add(procDir)
			rootDir.Add(resources_file)
			rootDir.Add(nodesdir)
			rootDir.Add(podsdir)
			time.Sleep(time.Second * 10)
		}
	}()

	return rootDir, nil
}

type resources struct {
	cpu    int64
	memory int64
}

//func (f *factory)
