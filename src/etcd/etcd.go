package etcd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"net"
	"strconv"
	"time"
	"unsafe"
)

func Register(etcd string, port int, weight int) error {
	fmt.Printf("Registering to etcd %s\n", etcd)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("Failed to create etcd client:", err)
		return err
	}
	addresses, _ := net.InterfaceAddrs()
	ip := ""
	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() && ipnet.IP.To4() != nil {
			ip = ipnet.IP.String()
		}
	}
	if ip == "" {
		fmt.Println("Failed to get IP address")
		return err
	}
	value := fmt.Sprintf("%s:%d|%d", ip, port, weight)
	lease := clientv3.NewLease(cli)
	grant, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		fmt.Println("Failed to grant lease from etcd:", err)
		return err
	}
	_, err = cli.Put(context.TODO(), "dubbomesh/"+value, value, clientv3.WithLease(grant.ID))
	if err != nil {
		fmt.Println("Failed to put to etcd:", err)
		return err
	}
	fmt.Println("Register success:", value)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			//fmt.Println("Heartbeat ...")
			_, err = lease.KeepAliveOnce(context.TODO(), grant.ID)
			if err != nil {
				fmt.Println("Failed to send heartbeat:", err)
			}
		}
	}()
	return nil
}

func Query(etcd string) ([]string, []int, error) {
	fmt.Printf("Querying from etcd %s\n", etcd)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("Failed to create etcd client:", err)
		return nil, nil, err
	}
	resp, err := cli.Get(context.TODO(), "dubbomesh/", clientv3.WithPrefix())
	if err != nil {
		fmt.Println("Failed to query etcd:", err)
		return nil, nil, err
	}
	fmt.Println("Providers:", resp.Kvs)
	servers := make([]string, len(resp.Kvs))
	weights := make([]int, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		sp := bytes.IndexByte(kv.Value, '|')
		if sp > 0 {
			host := kv.Value[:sp]
			weight := kv.Value[sp+1:]
			servers[i] = *(*string)(unsafe.Pointer(&host))
			weights[i], _ = strconv.Atoi(*(*string)(unsafe.Pointer(&weight)))
			if weights[i] <= 0 {
				weights[i] = 1
			}
		} else {
			servers[i] = *(*string)(unsafe.Pointer(&kv.Value))
			weights[i] = 1
		}
	}
	return servers, weights, nil
}
