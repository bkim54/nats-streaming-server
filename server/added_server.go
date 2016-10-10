package server

import (
	////"bufio"
	//"fmt"
	//"hash/crc32"
	////"io"
	//"os"
	//"path/filepath"
	//"time"
	//
	//"github.com/davecgh/go-spew/spew"
	//"github.com/bkim54/nats-streaming-server/stores"
	//"io/ioutil"
)


func (s *StanServer) DeleteChannel(name string) {
	// returns a channelStore
	if cs := s.store.LookupChannel(name); cs == nil {
		// does not exist
		Noticef(name + " does not exist")
		return;
	}
	s.store.DeleteChannel(name);
	s.store.Name()

}
