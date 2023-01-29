package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getFileOffset(t *testing.T) {
	type args struct {
		id uint64
	}
	testCase := []struct {
		name  string
		args  args
		want  uint64
		topic *Topic
	}{
		{
			name: "0 return 0",
			args: args{0},
			want: 0,
			topic: &Topic{
				msgBufferSize: 10,
			},
		},

		{
			name: "non buffer size value return the closest file offset",
			args: args{1},
			want: 0,
			topic: &Topic{
				msgBufferSize: 10,
			},
		},

		{
			name: "id bigger than buffer size",
			args: args{11},
			want: 10,
			topic: &Topic{
				msgBufferSize: 10,
			},
		},

		{
			name: "id bigger than buffer size",
			args: args{28},
			want: 20,
			topic: &Topic{
				msgBufferSize: 10,
			},
		},

		{
			name: "id equal to buffer size",
			args: args{10},
			want: 10,
			topic: &Topic{
				msgBufferSize: 10,
			},
		},

		{
			name: "id completely divisible by buffer size",
			args: args{20},
			want: 20,
			topic: &Topic{
				msgBufferSize: 10,
			},
		},

		{
			name: "handle error when msgBufferSize is 0",
			args: args{20},
			want: 0,
			topic: &Topic{
				msgBufferSize: 0,
			},
		},
	}

	for _, tt := range testCase {
		assert.EqualValues(t, tt.want, tt.topic.getFileOffset(tt.args.id), "getFileOffset: %s", tt.name)
	}
}
