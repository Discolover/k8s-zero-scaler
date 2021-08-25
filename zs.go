package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"io"
	"time"
	"context"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/go-redis/redis/v8"
)

var (
	clientset *kubernetes.Clientset
	server machinery.Server
	redisClient *redis.Client
)


func ZeroScaling(name, kind, namespace, uuid string) (err error) {
	key := fmt.Sprintf("%s:%s:%s", name, kind, namespace)
	effectiveUuid, err := redisClient.Get(context.TODO(), key).Result()
	if err != nil {
		return
	}
	if uuid != effectiveUuid {
		log.Printf("Skip task %s because effective task for %s:%s:%s" +
			"is %s\n", uuid, name, kind, namespace, effectiveUuid)
		return
	}
	switch kind {
	case "DaemonSet":
		err = fmt.Errorf("`DaemonSet`s are not zero scaled\n")
	case "Deployment":
		i := clientset.AppsV1().Deployments(namespace)
		d, err := i.Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			break
		}
		*d.Spec.Replicas = 0
		_, err = i.Update(context.TODO(), d, metav1.UpdateOptions{})
	case "ReplicaSet":
		i := clientset.AppsV1().ReplicaSets(namespace)
		r, err := i.Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			break
		}
		*r.Spec.Replicas = 0
		_, err = i.Update(context.TODO(), r, metav1.UpdateOptions{})
	case "StatefulSet":
		i := clientset.AppsV1().StatefulSets(namespace)
		s, err := i.Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			break
		}
		*s.Spec.Replicas = 0
		_, err = i.Update(context.TODO(), s, metav1.UpdateOptions{})
	default:
		err = fmt.Errorf("Not existing pod controller kind %s\n", kind)
	}

	return
}

func init() {
	// k8s-client related initializations
	k8sCnf, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalln(err)
	}

	clientset, err = kubernetes.NewForConfig(k8sCnf)
	if err != nil {
		log.Fatalln(err)
	}

	// machinery related initializations
	machineryCnf, err := config.NewFromEnvironment()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(machineryCnf) // @debug

	server, err := machinery.NewServer(machineryCnf)
	if err != nil {
		log.Fatalln(err)
	}

	worker := server.NewWorker("zero_scaler", 0)
	err = worker.Launch()
	if err != nil {
		log.Fatalln(err)
	}

	server.RegisterTask("zeroScaling", ZeroScaling)

	// redis-client related initializations
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func scheduleTask(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err) // @todo: It should not os.Exit(1)!!!
	}

	var a []string

	err = json.Unmarshal(raw, &a)
	if err != nil {
		log.Fatalln(err) // @todo: It should not os.Exit(1)!!!
	}

	log.Println(a) // @debug

	var args []tasks.Arg

	for _, v := range a {
		args = append(args, tasks.Arg{
			Type: "string",
			Value: v,
		})
	}

	log.Println(args) // @debug

	signature, err := tasks.NewSignature("zeroScaling", args)
	if err != nil {
		log.Fatalln(err) // @todo: It should not os.Exit(1)!!!
	}
	t := time.Now().UTC().Add(time.Minute)
	signature.ETA = &t

	// We want task to know its uuid to decide whether it is effective
	// pod controller in question.
	signature.Args = append(signature.Args, tasks.Arg{
		Type: "uint64",
		Value: signature.UUID,
	})


	key := fmt.Sprintf("%s:%s:%s", args[0], args[1], args[2]) // @todo: Is `Kind` really necessary?
	err = redisClient.Set(context.TODO(), key, signature.UUID, 0).Err()
	if err != nil {
		log.Fatalln(err) // @todo: It should not os.Exit(1)!!!
	}

	_, err = server.SendTask(signature) // @todo: Maybe, I should append result to some list
	if err != nil {
		log.Fatalln(err) // @todo: this. It should not os.Exit(1)!!!
	}

	log.Println(signature.Args) // @debug
}

func main() {
	http.HandleFunc("/", scheduleTask)
	//http.HandleFunc("/healthz", scheduleTask)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
