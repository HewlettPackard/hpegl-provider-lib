// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

//nolint:scopelint
package atf

import (
	"strings"
	"testing"
)

func Test_path(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
		want string
	}{
		{
			name: "2 value",
			args: []interface{}{"acc", 0},
			want: "acc.0",
		},
		{
			name: "1 value",
			args: []interface{}{"acc"},
			want: "acc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := path(tt.args...); got != tt.want {
				t.Errorf("path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLocalName(t *testing.T) {
	type args struct {
		res string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test case 1: hpegl_vmaas_instance",
			args: args{
				res: "hpegl_vmaas_instance",
			},
			want: "instance",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLocalName(tt.args.res); got != tt.want {
				t.Errorf("getLocalName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTag(t *testing.T) {
	type args struct {
		isResource bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test case 1- resource",
			args: args{
				isResource: true,
			},
			want: "resources",
		},
		{
			name: "Test case 2- data source",
			args: args{
				isResource: false,
			},
			want: "data-sources",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTag(tt.args.isResource); got != tt.want {
				t.Errorf("getTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getType(t *testing.T) {
	type args struct {
		isResource bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test case 1, resource",
			args: args{
				isResource: true,
			},
			want: "resource",
		},
		{
			name: "Test case 2, data source",
			args: args{
				isResource: false,
			},
			want: "data",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getType(tt.args.isResource); got != tt.want {
				t.Errorf("getType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toInt(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Test 1",
			args: args{
				str: "12",
			},
			want: 12,
		},
		{
			name: "Test 2, invalid data",
			args: args{
				str: "abc",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toInt(tt.args.str); got != tt.want {
				t.Errorf("toInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFrame(t *testing.T) {
	type args struct {
		skipFrames int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test case 1",
			args: args{
				skipFrames: 10,
			},
			want: "Test_getFrame",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFrame(tt.args.skipFrames); !strings.Contains(got, tt.want) {
				t.Errorf("getFrame() = %v, should cotains %v", got, tt.want)
			}
		})
	}
}
