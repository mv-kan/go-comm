## Example 
```
package main

import (
	"fmt"

	gocomm "github.com/mv-kan/go-comm"
)

func main() {
	fmt.Println("I am writer")
	p, err := gocomm.NewPort("/dev/", 115200, 0)
	if err != nil {
		panic(err)
	}
	input := make(chan string)
	conn, msgChan, err := gocomm.NewConnection(p, input, 0, "\n", "\n")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	msg := <-msgChan
	fmt.Println("msg.Data=", msg.Data)
}
```