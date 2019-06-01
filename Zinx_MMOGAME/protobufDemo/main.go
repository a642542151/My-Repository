
package main

import (
	"Project03/protobufDemo/pb"
	"fmt"
	"github.com/golang/protobuf/proto"
)

func main() {
	//定义一个protobuf结构体对象
	person:= &pb.Person{
		Name:"jack",
		Age:16,
		Emails: []string{"danbing.at@gmail.com", "danbing_at@163.com"},
		Phones: []*pb.PhoneNumber{
			&pb.PhoneNumber{
				Number:"1111111",
			},
			&pb.PhoneNumber{
				Number:"22222222",
			},
			&pb.PhoneNumber{
				Number:"33333333",
			},
		},
		//oneof赋值
		Data: &pb.Person_Socre{
			Socre:100,
		},
	}

	//将一个protobuf结构体对象 转化成二进制数据
	//任何proto message结构体 在go中他们都是基础Message接口的

	//编码
	data , err := proto.Marshal(person)
	if err != nil {
		fmt.Println("marshal err ", err)
		return
	}

	//data就是我们要刚给对端发送的二进制数据

	//对端已经收到了data了
	//解码
	newPerson := &pb.Person{}
	err = proto.Unmarshal(data, newPerson) //将data解码值 newPerson结构体中
	if err != nil {
		fmt.Println("unmarshal err ", err)
		return
	}

	fmt.Println("源数据：" , person)
	fmt.Println("解码之后数据:", newPerson)
	fmt.Println("name = ", newPerson.GetName(),"age = ",newPerson.GetAge(), " emails: ", newPerson.GetEmails())
	fmt.Println("numbers = ", newPerson.GetPhones())
	fmt.Println("score = ", newPerson.GetSocre())
}
