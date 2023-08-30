package models

import "time"

type Order struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement"`
	UserId      string    `gorm:"column:user_id;not null;size:64"`
	Username    string    `gorm:"column:username;not null;size:32"`
	Energy      int64     `gorm:"column:energy;default:0"`          //兑换能量
	Payments    string    `gorm:"column:payments;default:0.0"`      //实际所需金额
	Amount      float64   `gorm:"column:amount;default:0.0"`        //订单所需金额
	ToAddress   string    `gorm:"column:to_address;size:64"`        //租用能量接收地址
	Duration    string    `gorm:"column:duration;not null;size:32"` //能量租用时长
	Status      int8      `gorm:"column:status;size:8;default:0"`   //订单状态
	Blocks      int64     `gorm:"column:blocks;not null"`           //goodies rent token blocks
	FeesAddress string    `gorm:"column:fees_address;size:64"`      //goodies rent fees address
	Multiplier  float64   `gorm:"column:multiplier;default:1.0"`    //goodies rent energy multiplier
	MessageId   int       `gorm:"column:message_id;default:0"`
	ChatId      string    `gorm:"column:chat_id;size:64;default:null"`
	Finished    bool      `gorm:"column:finished;default:false"`
	Expired     bool      `gorm:"column:expired;default:false"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (o *Order) TableName() string {
	return "tb_order"
}
