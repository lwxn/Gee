package gee

import (
	"fmt"
	"log"
	"time"
)

func Logger() HandlerFunc{
	fmt.Println("----------------------logging-----------------")
	return func(c *Context){
		t := time.Now()

		c.Next()

		log.Printf("[%d] %s in %v",c.StatusCode,c.Req.RequestURI,time.Since(t))
	}
}
