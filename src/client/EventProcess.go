package main

import (
	"actor"
	"message"
	"github.com/golang/protobuf/proto"
	"fmt"
	"base"
	"server/world/game/lmath"
	"network"
)

type (
	EventProcess struct {
		actor.Actor

		Client *network.ClientSocket
		AccountId int64
		PlayerId int64
		AccountName string
		SimId int64
		Pos lmath.Point3F
		Rot lmath.Point3F
	}

	IEventProcess interface {
		actor.IActor
		LoginGame()
		LoginAccount()
		SendPacket(proto.Message)
	}
)

func SendPacket(packet proto.Message){
	buff := message.Encode(packet)
	buff = base.SetTcpEnd(buff)
	CLIENT.Send(buff)
}

func (this *EventProcess) SendPacket(packet proto.Message){
	buff := message.Encode(packet)
	buff = base.SetTcpEnd(buff)
	this.Client.Send(buff)
}

func (this *EventProcess) PacketFunc(socketid int, buff []byte) bool {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("EventProcess PacketFunc", err)
		}
	}()

	packetId, data := message.Decode(buff)
	packet := message.GetPakcet(packetId)
	if packet == nil{
		return true
	}
	err := proto.Unmarshal(data, packet)
	if err == nil{
		bitstream := base.NewBitStream(make([]byte, 1024), 1024)
		if !message.GetProtoBufPacket(packet, bitstream) {
			return true
		}
		var io actor.CallIO
		io.Buff = bitstream.GetBuffer()
		io.SocketId = socketid
		this.Send(io)
		return true
	}

	return true
}

func (this *EventProcess) Init(num int) {
	this.Actor.Init(num)
	this.Pos = lmath.Point3F{1, 1, 1}
	this.RegisterCall("W_C_SelectPlayerResponse", func(packet *message.W_C_SelectPlayerResponse) {
		this.AccountId = *packet.AccountId
		nLen := len(packet.PlayerData)
		//fmt.Println(len(packet.PlayerData), this.AccountId, packet.PlayerData)
		if nLen == 0{
			packet1 := &message.C_W_CreatePlayerRequest{PacketHead:message.BuildPacketHead( this.AccountId, int(message.SERVICE_WORLDSERVER)),
				PlayerName:proto.String("我是大坏蛋"),
				Sex:proto.Int32(int32(0)),}
			this.SendPacket(packet1)
		}else{
			this.PlayerId = *packet.PlayerData[0].PlayerID
			this.LoginGame()
		}
	})

	this.RegisterCall("W_C_CreatePlayerResponse", func(packet *message.W_C_CreatePlayerResponse) {
		if *packet.Error == 0 {
			this.PlayerId = *packet.PlayerId
			this.LoginGame()
		}else{//创建失败

		}
	})

	this.RegisterCall("A_C_LoginRequest", func(packet *message.A_C_LoginRequest) {
		if *packet.Error == base.ACCOUNT_NOEXIST {
			packet1 := &message.C_A_RegisterRequest{PacketHead:message.BuildPacketHead( 0, int(message.SERVICE_ACCOUNTSERVER)),
				AccountName:packet.AccountName, SocketId: proto.Int32(0)}
			this.SendPacket(packet1)
		}
	})

	this.RegisterCall("A_C_RegisterResponse", func(packet *message.A_C_RegisterResponse) {
		//注册失败
		if *packet.Error != 0 {
		}
	})

	this.RegisterCall("W_C_ChatMessage", func(packet *message.W_C_ChatMessage) {
		fmt.Println("收到【", *packet.RecverName, "】发送的消息[", *packet.Message+"]")
	})

	this.RegisterCall("W_C_LoginMap", func(packet *message.W_C_LoginMap) {
		this.SimId = *packet.Id
		this.Pos = lmath.Point3F{*packet.Pos.X, *packet.Pos.Y, *packet.Pos.Z}
		this.Rot = lmath.Point3F{0, 0, *packet.Rotation}
		fmt.Println("login map")
	})

	this.RegisterCall("W_C_Move", func(packet *message.W_C_Move) {
		if this.SimId == *packet.Id{
			this.Pos = lmath.Point3F{*packet.Pos.X, *packet.Pos.Y, *packet.Pos.Z}
			this.Rot = lmath.Point3F{0, 0, *packet.Rotation}
		}else{
			fmt.Printf("simobj:[%d], Pos:[x:%d, y:%d, z:%d], Rot[%d]", *packet.Id, *packet.Pos.X, *packet.Pos.Y, *packet.Pos.Z, *packet.Rotation)
		}
		//this.Move(0, 100.0)
	})
	this.Actor.Start()
}

func (this *EventProcess)  LoginGame(){
	packet1 := &message.C_W_Game_LoginRequset{PacketHead:message.BuildPacketHead( this.AccountId, int(message.SERVICE_WORLDSERVER)),
		PlayerId:proto.Int64(this.PlayerId),}
	this.SendPacket(packet1)
}

var(
	id int
)

func (this *EventProcess)  LoginAccount() {
	id++
	//this.AccountName = fmt.Sprintf("test%d", id)
	this.AccountName = fmt.Sprintf("test%d", base.RAND().RandI(0, 7000))
	packet1 := &message.C_A_LoginRequest{PacketHead: message.BuildPacketHead(0, int(message.SERVICE_ACCOUNTSERVER)),
		AccountName: proto.String(this.AccountName), BuildNo: proto.String(base.BUILD_NO), SocketId: proto.Int32(0)}
	this.SendPacket(packet1)
}

var(
	PACKET *EventProcess
)

func (this *EventProcess)  Move(yaw float32, time float32) {
	packet1 := &message.C_W_Move{PacketHead: message.BuildPacketHead(this.AccountId, int(message.SERVICE_WORLDSERVER)),
		Move: &message.C_W_Move_Move{Mode: proto.Int32(0), Normal:&message.C_W_Move_Move_Normal{Pos:&message.Point3F{X:proto.Float32(this.Pos.X), Y:proto.Float32(this.Pos.Y), Z:proto.Float32(this.Pos.Z)}, Yaw:proto.Float32(yaw), Duration:proto.Float32(time)}}}
	this.SendPacket(packet1)
}