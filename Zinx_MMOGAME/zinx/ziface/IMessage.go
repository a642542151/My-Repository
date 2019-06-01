
package ziface

type IMessage interface {
	//getter
	GetMsgId() uint32
	GetMsgLen() uint32
	GetMsgData() []byte

	//setter
	SetMsgId(uint32)
	SetData([]byte)
	SetDatalen(uint32)
}
