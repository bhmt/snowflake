package snowflake_test

import (
	"snowflake"
	"testing"
)

func TestParseId(t *testing.T) {
	//                                     timestamp      | worker id | sequence
	// 0000 0000 0000 0000 0000 0000 0000 0000 0000 0000 0100 0000 0001 0000 0000 0001
	tests := []struct {
		name      string
		id        int64
		timestamp int64
		workerId  int16
		sequence  int16
	}{
		{
			name:      "sequence",
			id:        1,
			timestamp: 0,
			workerId:  0,
			sequence:  1,
		},
		{
			name:      "worker sequence",
			id:        0x1001,
			timestamp: 0,
			workerId:  1,
			sequence:  1,
		},
		{
			name:      "all",
			id:        0x401001,
			timestamp: 1,
			workerId:  1,
			sequence:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			obtained := snowflake.ParseId(test.id)
			expected := snowflake.Id{
				Id:        test.id,
				Timestamp: test.timestamp,
				WorkerId:  test.workerId,
				Sequence:  test.sequence,
			}

			if obtained != expected {
				t.Fatalf("unexpected value, want %#v, got %#v", expected, obtained)
			}
		})
	}
}

func TestNextId(t *testing.T) {
	sf, _ := snowflake.NewSnowflake(1)
	current_id := sf.NextId()
	sid := snowflake.ParseId(current_id)

	current_timestamp := sid.Timestamp
	current_sequence := sid.Sequence

	for i := 0; i < snowflake.BitMaskSequence*10; i++ {
		id := sf.NextId()
		sid = snowflake.ParseId(id)

		if sid.Sequence == 0 && sid.Timestamp <= current_timestamp {
			t.Fatal("time must increase for new sequence count")
		}

		if current_timestamp == sid.Timestamp && sid.Sequence <= current_sequence {
			t.Fatal("sequence must increase for unchanged timestamp")
		}

		if id <= current_id {
			t.Fatal("id not increasing")
		}

		current_id = id
		current_timestamp = sid.Timestamp
		current_sequence = sid.Sequence
	}
}
