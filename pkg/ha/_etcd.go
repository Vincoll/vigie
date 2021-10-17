package ha

import (
	"context"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"time"
)

// https://medium.com/@felipedutratine/leader-election-in-go-with-etcd-2ca8f3876d79
// https://programmer.help/blogs/a-concise-tutorial-of-golang-etcd.html

// https://www.compose.com/articles/utilizing-etcd3-with-go/
// https://github.com/the-gigi/go-etcd3-demo/blob/master/main.go
func main() {
	var name = flag.String("name", "", "give a name")
	flag.Parse() // Create a etcd client
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close() // create a sessions to elect a Leader
	s, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()
	e := concurrency.NewElection(s, "/leader-election/")
	ctx := context.Background() // Elect a leader (or wait that the leader resign)

	if err := e.Campaign(ctx, "e"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("leader election for ", *name)
	fmt.Println("Do some work in", *name)
	time.Sleep(5 * time.Second)

	if err := e.Resign(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("resign ", *name)
}
