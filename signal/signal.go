package signal

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

//https://github.com/kubernetes/kubernetes/blob/597a684bb0653a607009e97299986763ebfde020/staging/src/k8s.io/apiserver/pkg/server/signal.go#L33

var once sync.Once
var stopCh = make(chan struct{})

func SetupStopCh() chan struct{} {
	once.Do(func() {
		signals := make(chan os.Signal, 2)
		signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			signal := <-signals
			fmt.Printf("receive signal: %s\n", signal)
			close(stopCh)
			signal = <-signals
			fmt.Printf("receive signal: %s, exit\n", signal)
			os.Exit(1)
		}()
	})
	return stopCh
}
