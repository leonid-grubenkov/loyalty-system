package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/leonid-grubenkov/loyalty-system/internal/models"
)

var cliet http.Client

const posturl = "http://localhost:8090/api/orders"

func (s *Service) Worker(id int, postUrl string, orders <-chan int) {
	for order := range orders {
		log.Println("worker", id, "start order", order)
		for {
			resOrder, err := getAccrual(order, postUrl)
			if err != nil {
				log.Println(err)
				break
			}

			switch resOrder.Status {
			case "REGISTERED":
				log.Println(order, " - ", resOrder.Status)
				err := s.db.ChangeStatus(context.Background(), order, resOrder.Status)
				if err != nil {
					log.Println("error status registred - ", err)
				}
			case "INVALID":
				log.Println(order, " - ", resOrder.Status)
				err := s.db.ChangeStatus(context.Background(), order, resOrder.Status)
				if err != nil {
					log.Println("error status invalid - ", err)
				}
				break
			case "PROCESSING":
				log.Println(order, " - ", resOrder.Status)
				err := s.db.ChangeStatus(context.Background(), order, resOrder.Status)
				if err != nil {
					log.Println("error status processing - ", err)
				}
			case "PROCESSED":
				log.Println(order, " - ", resOrder.Status)
				err := s.db.ChangeAccrual(context.Background(), order, resOrder.Status, resOrder.Accrual)
				if err != nil {
					log.Println("error status processed changeaccrual - ", err)
				}
				err = s.db.AddBallance(context.Background(), resOrder.Login, resOrder.Accrual)
				if err != nil {
					log.Println("error status processed addbalance - ", err)
				}
				break
			default:
				log.Println(order, " - ", resOrder.Status)
			}
		}
		log.Println("worker", id, "finish order ", order)
	}
}

func getAccrual(order int, postUrl string) (*models.Order, error) {
	for {
		reqUrl := fmt.Sprint(postUrl, "/api/orders/", order)
		log.Println("order", order, "URL - ", reqUrl)
		r, err := http.NewRequest("GET", reqUrl, nil)
		if err != nil {
			log.Printf("Error on method newRequest get orders - %s", err)
		}
		client := http.Client{}
		res, err := client.Do(r)
		if err != nil {
			log.Printf("Error on method do get orders - %s", err)
		}

		if res != nil {
			defer res.Body.Close()

			switch res.StatusCode {
			case http.StatusOK:
				log.Println(order, "accrual service result OK")
				var buf bytes.Buffer
				var order models.Order

				_, err := buf.ReadFrom(res.Body)
				if err != nil {
					return nil, err
				}

				if err = json.Unmarshal(buf.Bytes(), &order); err != nil {
					return nil, err
				}

			case http.StatusNoContent:
				log.Println(order, "not register")
			case http.StatusTooManyRequests:
				log.Println(order, "too many requests")
			case http.StatusInternalServerError:
				log.Println(order, "internal server error")
			default:
				log.Println("unknown status: ", res.Status)
			}
		}

		time.Sleep(5 * time.Second)

	}
}