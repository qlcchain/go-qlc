package p2p

import (
	"reflect"
	"testing"
)

func getMockQlcMessage1() *QlcMessage {
	data := "testMessage"
	msgType := byte(0x01)
	version := byte(0x01)
	content := NewQlcMessage([]byte(data), version, MessageType(msgType))
	qlcMsg := &QlcMessage{
		content:     content,
		messageType: MessageType(msgType),
	}
	return qlcMsg
}

func TestQlcMessage_MagicNumber(t *testing.T) {
	qm := getMockQlcMessage1()
	magicnumber := []byte{0x51, 0x4C, 0x43}
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, magicnumber},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if got := message.MagicNumber(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QlcMessage.MagicNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcMessage_Version(t *testing.T) {
	qm := getMockQlcMessage1()
	version := byte(0x01)
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name   string
		fields fields
		want   byte
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, version},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if got := message.Version(); got != tt.want {
				t.Errorf("QlcMessage.Version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcMessage_MessageType(t *testing.T) {
	qm := getMockQlcMessage1()
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name   string
		fields fields
		want   MessageType
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, qm.messageType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if got := message.MessageType(); got != tt.want {
				t.Errorf("QlcMessage.MessageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcMessage_MessageData(t *testing.T) {
	qm := getMockQlcMessage1()
	data := "testMessage"
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, []byte(data)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if got := message.MessageData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QlcMessage.MessageData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcMessage_DataLength(t *testing.T) {
	var datalen uint32
	qm := getMockQlcMessage1()
	datalen = 11
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, datalen},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if got := message.DataLength(); got != tt.want {
				t.Errorf("QlcMessage.DataLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcMessage_DataCheckSum(t *testing.T) {
	var checksum uint32
	qm := getMockQlcMessage1()
	checksum = 401271118
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, checksum},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if got := message.DataCheckSum(); got != tt.want {
				t.Errorf("QlcMessage.DataCheckSum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcMessage_HeaderData(t *testing.T) {
	qm := getMockQlcMessage1()
	headdata := []byte{81, 76, 67, 1, 1, 0, 0, 0, 11}
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, headdata},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if got := message.HeaderData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QlcMessage.HeaderData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewQlcMessage(t *testing.T) {
	qm := getMockQlcMessage1()
	data := "testMessage"
	msgType := byte(0x01)
	version := byte(0x01)
	type args struct {
		data           []byte
		currentVersion byte
		messageType    MessageType
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
		{"OK", args{[]byte(data), version, MessageType(msgType)}, qm.content},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQlcMessage(tt.args.data, tt.args.currentVersion, tt.args.messageType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQlcMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseQlcMessage(t *testing.T) {
	data1 := []byte{1, 2}
	qm := getMockQlcMessage1()
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *QlcMessage
		wantErr bool
	}{
		// TODO: Add test cases.
		{"BadDatalen", args{data1}, nil, true},
		{"OK", args{qm.content}, qm, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseQlcMessage(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQlcMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ParseQlcMessage() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestQlcMessage_ParseMessageData(t *testing.T) {
	data1 := []byte{1, 2}
	qm := getMockQlcMessage1()

	type fields struct {
		content     []byte
		messageType MessageType
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"BadDatalen", fields{qm.content, qm.messageType}, args{data1}, true},
		{"BadChecksum", fields{qm.content, qm.messageType}, args{qm.content}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if err := message.ParseMessageData(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("QlcMessage.ParseMessageData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcMessage_VerifyHeader(t *testing.T) {
	qm := getMockQlcMessage1()
	qm2 := getMockQlcMessage1()
	qm2.content[2] = byte(0x01)
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, false},
		{"badheader", fields{qm2.content, qm2.messageType}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if err := message.VerifyHeader(); (err != nil) != tt.wantErr {
				t.Errorf("QlcMessage.VerifyHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcMessage_VerifyData(t *testing.T) {
	qm := getMockQlcMessage1()
	qm2 := getMockQlcMessage1()
	qm2.content[len(qm2.content)-1] = byte(0x01)
	type fields struct {
		content     []byte
		messageType MessageType
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"OK", fields{qm.content, qm.messageType}, false},
		{"badchecksum", fields{qm2.content, qm2.messageType}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &QlcMessage{
				content:     tt.fields.content,
				messageType: tt.fields.messageType,
			}
			if err := message.VerifyData(); (err != nil) != tt.wantErr {
				t.Errorf("QlcMessage.VerifyData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFromUint32(t *testing.T) {
	data := "test"
	dataint := 0x74657374
	type args struct {
		v uint32
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
		{"OK", args{uint32(dataint)}, []byte(data)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromUint32(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint32(t *testing.T) {
	data := "test"
	dataint := 0x74657374
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// TODO: Add test cases.
		{"OK", args{[]byte(data)}, uint32(dataint)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uint32(tt.args.data); got != tt.want {
				t.Errorf("Uint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	a1 := "test1"
	a2 := "test2"
	a3 := "test1"
	type args struct {
		a []byte
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{"OK", args{[]byte(a1), []byte(a3)}, true},
		{"OK", args{[]byte(a1), []byte(a2)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
