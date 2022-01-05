// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package rangeplugin

import (
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var leasefile string = `02:00:00:00:00:00 10.0.0.0 2000-01-01T00:00:00Z 
02:00:00:00:00:01 10.0.0.1 2000-01-01T00:00:00Z host1
02:00:00:00:00:02 10.0.0.2 2000-01-01T00:00:00Z host2
02:00:00:00:00:03 10.0.0.3 2000-01-01T00:00:00Z host3
02:00:00:00:00:04 10.0.0.4 2000-01-01T00:00:00Z host4
02:00:00:00:00:05 10.0.0.5 2000-01-01T00:00:00Z host5
`

var expire = time.Date(2000, 01, 01, 00, 00, 00, 00, time.UTC)
var records = []struct {
	mac string
	ip  *Record
}{
	{mac: "02:00:00:00:00:00", ip: &Record{IP: net.IPv4(10, 0, 0, 0), expires: expire}},
	{mac: "02:00:00:00:00:01", ip: &Record{Hostname: "host1", IP: net.IPv4(10, 0, 0, 1), expires: expire}},
	{mac: "02:00:00:00:00:02", ip: &Record{Hostname: "host2", IP: net.IPv4(10, 0, 0, 2), expires: expire}},
	{mac: "02:00:00:00:00:03", ip: &Record{Hostname: "host3", IP: net.IPv4(10, 0, 0, 3), expires: expire}},
	{mac: "02:00:00:00:00:04", ip: &Record{Hostname: "host4", IP: net.IPv4(10, 0, 0, 4), expires: expire}},
	{mac: "02:00:00:00:00:05", ip: &Record{Hostname: "host5", IP: net.IPv4(10, 0, 0, 5), expires: expire}},
}

func TestLoadRecords(t *testing.T) {
	parsedRec, err := loadRecords(strings.NewReader(leasefile))
	if err != nil {
		t.Fatalf("Failed to load records from file: %v", err)
	}

	mapRec := make(map[string]*Record)
	for _, rec := range records {
		mapRec[rec.mac] = rec.ip
	}

	assert.Equal(t, mapRec, parsedRec, "Loaded records differ from what's in the file")
}

func TestWriteRecords(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "coredhcptest")
	if err != nil {
		t.Skipf("Could not setup file-based test: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	pl := PluginState{
		Recordsv4: make(map[string]*Record),
	}
	if err := pl.registerBackingFile(tmpfile.Name()); err != nil {
		t.Fatalf("Could not setup file")
	}
	defer pl.leasefile.Close()

	for _, rec := range records {
		pl.Recordsv4[rec.mac] = rec.ip
	}
	if err := pl.saveState(); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	written, err := ioutil.ReadAll(tmpfile)
	if err != nil {
		t.Fatalf("Could not read back temp file")
	}
	assert.Equal(t, leasefile, string(written), "Data written to the file doesn't match records")

	if err := pl.saveState(); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	written, err = ioutil.ReadAll(tmpfile)
	if err != nil {
		t.Fatalf("Could not read back temp file")
	}
	assert.Equal(t, leasefile, string(written), "Data rewritten to the file doesn't match records")
}
