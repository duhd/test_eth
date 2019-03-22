package utils

import (
  "os"
  "fmt"
  "io/ioutil"
  "strings"
  "path/filepath"
  "encoding/json"
  "time"
  "github.com/go-redis/redis"
)

type Transaction struct {
        Id                string  `json:"Id"`
        RequestTime       int64   `json:"RequestTime"`
        TxReceiveTime     int64   `json:"TxReceiveTime"`
        TxConfirmedTime    []int64 `json:"TxConfiredTime"`
   }

var Redis_client *redis.Client


func LoadKeyStores(root string){
    var files []string
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
               files = append(files, path)
               return nil
           })
    if err != nil {
         panic(err)
    }
    for _, file := range files {
         fmt.Println("File:", file)
         list := strings.Split(file,"--")
         if len(list) == 3 {
             account := "account:" + list[2]
             keyjson, err := ioutil.ReadFile(file)
             if err != nil {
                  fmt.Println("Error in read file: ", file )
                  continue
             }
             //Set key in redis
              err = Redis_client.Set(account,string(keyjson), 0).Err()
              if err != nil {
                panic(err)
              }
         }
    }
}

func LogStart(key string,requesttime int64){
  trans :=  &Transaction{
              Id: key,
              RequestTime: requesttime,
              TxReceiveTime: time.Now().UnixNano()}
  value, err := json.Marshal(trans)
  if err != nil {
      fmt.Println(err)
      return
  }
  err = Redis_client.Set("transaction:" + key,string(value), 0).Err()
	if err != nil {
		panic(err)
	}
}

func LogEnd(key string){
      val, err2 := Redis_client.Get("transaction:" + key).Result()
      if err2 != nil {
          fmt.Println("Cannot find transaction: ", key)
          return
      }
      data := &Transaction{}
      err := json.Unmarshal([]byte(val), data)
      if err != nil {
          fmt.Println("Cannot parse data ", err)
          return
      }
      data.TxConfirmedTime = append(data.TxConfirmedTime, time.Now().UnixNano())
      value, err := json.Marshal(data)

      err = Redis_client.Set("transaction:" + key,string(value), 0).Err()
    	if err != nil {
    	     fmt.Println("Cannot set data ", err)
    	}
      fmt.Println("Finish write transaction: ", key)
}