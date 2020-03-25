package pkg

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	schedulercache "test.io/scheduler-utils/cache"
)
var(

	kubeconfig  string = "/opt/eprockube/kubeconfig.yaml"

	schema *kInfo
	
)

type kInfo struct {
	client kubernetes.Interface
	db schedulercache.Cache
	//expired time.Time
	podLister  v1lister.PodLister
}


func Init()*kInfo{

	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		logrus.Error(err)
		return nil
	}
	kubeConfig.QPS = 50
	kubeConfig.Burst = 60
	kubeConfig.ContentType="application/vnd.kubernetes.protobuf"
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	ver,_:=client.ServerVersion()

	fmt.Println(ver.String())
	
	stop:=make(chan struct{})

	sf := informers.NewSharedInformerFactory(client, time.Minute*10)
	
	cacher:=schedulercache.New(time.Second*100,stop)
	sf.Core().V1().Nodes().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*v1.Node)
			if !ok {
				klog.Errorf("cannot convert to *v1.Node: %v", obj)
				return
			}

			if err := cacher.AddNode(node); err != nil {
				klog.Errorf("scheduler cache AddNode failed: %v", err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldNode, ok := oldObj.(*v1.Node)
			if !ok {
				klog.Errorf("cannot convert oldObj to *v1.Node: %v", oldObj)
				return
			}
			newNode, ok := newObj.(*v1.Node)
			if !ok {
				klog.Errorf("cannot convert newObj to *v1.Node: %v", newObj)
				return
			}
			cacher.UpdateNode(oldNode,newNode)
		},
		DeleteFunc: func(obj interface{}) {
			var node *v1.Node
			switch t := obj.(type) {
			case *v1.Node:
				node = t
			case cache.DeletedFinalStateUnknown:
				var ok bool
				node, ok = t.Obj.(*v1.Node)
				if !ok {
					klog.Errorf("cannot convert to *v1.Node: %v", t.Obj)
					return
				}
			default:
				klog.Errorf("cannot convert to *v1.Node: %v", t)
				return
			}
			klog.V(3).Infof("delete event for node %q", node.Name)
			cacher.RemoveNode(node)
		},
	})

	sf.Core().V1().Pods().Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *v1.Pod:
				return assignedNonTerminatedPod(t)
			case cache.DeletedFinalStateUnknown:
				if pod, ok := t.Obj.(*v1.Pod); ok {
					return assignedNonTerminatedPod(pod)
				}
				return false
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if !ok {
					klog.Errorf("cannot convert to *v1.Pod: %v", obj)
					return
				}
				klog.V(3).Infof("add event for scheduled pod %s/%s ", pod.Namespace, pod.Name)

				if err := cacher.AddPod(pod); err != nil {
					klog.Errorf("scheduler cache AddPod failed: %v", err)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldPod, ok := oldObj.(*v1.Pod)
				if !ok {
					klog.Errorf("cannot convert oldObj to *v1.Pod: %v", oldObj)
					return
				}
				newPod, ok := newObj.(*v1.Pod)
				if !ok {
					klog.Errorf("cannot convert newObj to *v1.Pod: %v", newObj)
					return
				}
				cacher.UpdatePod(oldPod,newPod)
			},
			DeleteFunc: func(obj interface{}) {
				var pod *v1.Pod
				switch t := obj.(type) {
				case *v1.Pod:
					pod = t
				case cache.DeletedFinalStateUnknown:
					var ok bool
					pod, ok = t.Obj.(*v1.Pod)
					if !ok {
						klog.Errorf("cannot convert to *v1.Pod: %v", t.Obj)
						return
					}
				default:
					klog.Errorf("cannot convert to *v1.Pod: %v", t)
					return
				}
				klog.V(3).Infof("delete event for scheduled pod %s/%s ", pod.Namespace, pod.Name)
				cacher.RemovePod(pod)
			},
		},
	})
	go sf.Core().V1().Pods().Informer().Run(stop)
	go sf.Core().V1().Nodes().Informer().Run(stop)
	for{
		time.Sleep(time.Second)
		if sf.Core().V1().Pods().Informer().HasSynced()&&sf.Core().V1().Nodes().Informer().HasSynced(){
			break
		}
	}

	schema=&kInfo{client,cacher,sf.Core().V1().Pods().Lister()}
	return schema
}


func assignedNonTerminatedPod(pod *v1.Pod) bool {
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return false
	}

	//if pod.DeletionTimestamp != nil {
	//	return false
	//}

	// pending pod
	if len(pod.Spec.NodeName)==0{
		return false
	}
	return true
}
