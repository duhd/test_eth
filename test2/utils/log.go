package utils
import (
  "fmt"
  "encoding/json"
  "time"
)
func  LogStart(key string, nonce uint64, requesttime int64) bool {
  client := Rclients.getClient()
  trans :=  &Transaction{
              Id: key,
              TxNonce: nonce,
              RequestTime: requesttime,
              TxReceiveTime: time.Now().UnixNano()}
  value, err := json.Marshal(trans)
  if err != nil {
      fmt.Println(err)
      return false
  }
  err = client.Set("transaction:" + key,string(value), 0).Err()
  if err != nil {
    fmt.Println(time.Now()," Write transaction to redis error: ", err)
    return false
  }
  return true
}

func  LogEnd(key string, nonce uint64){
      client := Rclients.getClient()
      val, err2 := client.Get("transaction:" + key).Result()
      if err2 != nil {
          fmt.Println(time.Now()," Cannot find transaction: ", key)
          return
      }

      data := &Transaction{}
      err := json.Unmarshal([]byte(val), data)
      if err != nil {
          fmt.Println(time.Now()," Cannot parse data ", err)
          return
      }

      TxConfirmedTime := time.Now().UnixNano()
      data.TxConfirmedTime = append(data.TxConfirmedTime,TxConfirmedTime )
      value, err := json.Marshal(data)

      err = client.Set("transaction:" + key,string(value), 0).Err()
    	if err != nil {
        fmt.Println(time.Now()," err: Cannot set data - ", err)
    	}

    time_receive_ms := (data.TxReceiveTime - data.RequestTime)/1000000
  time_confirm_ms := (TxConfirmedTime  - data.RequestTime )/1000000
  
      if data.TxNonce != nonce {
        fmt.Println(time.Now()," nonce:",data.TxNonce," tx:",key," request:",data.RequestTime," receive:", time_receive_ms, " error:",nonce)
      }


  fmt.Println(time.Now()," nonce:",data.TxNonce," tx:",key," request:",data.RequestTime," receive:", time_receive_ms, " confirm:",time_confirm_ms)
}

func Report() string {
      client := Rclients.getClient()
      keys, err  := client.Keys("transaction:*").Result()
      if err != nil {
        // handle error
        fmt.Println(time.Now()," Cannot get keys ")
      }
      vals, err1 := client.MGet(keys...).Result()
      if err1 != nil {
        // handle error
        fmt.Println(time.Now()," Cannot get values of  keys: ", keys)
      }

      fmt.Println("Elements: ", len(keys))
      diff_arr1 := []int64{}
      diff_arr := []int64{}

      for _, element := range vals {
          data := &Transaction{}
          err2 := json.Unmarshal([]byte(element.(string)), data)
          if err2 != nil {
              fmt.Println(time.Now()," Element:", element, ", Error:", err2)
              continue
          }
          fmt.Println("ID:",data.Id,"RequestTime:",data.RequestTime,
            "TxReceiveTime:",data.TxReceiveTime,"TxConfirmedTime:",data.TxConfirmedTime)

          var max int64 = 0
          if data.TxConfirmedTime != nil {
              for _,value := range data.TxConfirmedTime {
                  if value > max {
                     max = value
                  }
              }
              diff1 := data.TxReceiveTime - data.RequestTime
              diff_arr1 = append(diff_arr1,diff1)
          }
          // else {
          //     max = time.Now().UnixNano()
          // }
          if max >0 {
              diff := max  - data.TxReceiveTime
              diff_arr = append(diff_arr,diff)
          }
      }
      var total1 int64 = 0
    	for _, value1:= range diff_arr1 {
    		total1 += value1
    	}
      len1 := int64(len(diff_arr1))
      var avg1 int64 = 0
      if len1 >0 {
        	avg1 = total1/(len1 *1000)
      }

      var total int64 = 0
    	for _, value:= range diff_arr {
    		total += value
    	}
      len2 := int64(len(keys))
      len := int64(len(diff_arr))
      var avg int64 = 0
      if len >0 {
        	avg = total/(len *1000)
      }
      return fmt.Sprintf("Total Tx: %v , Total Complete TX: %v ,Avg RequestTime: %v , Avg Onchain: %v ", len2, len,avg1, avg)
}
