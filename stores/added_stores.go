package stores

import (
	//"bufio"
	"fmt"
	"hash/crc32"
	//"io"
	"os"
	"path/filepath"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/bkim54/nats-streaming-server/spb"
	"io/ioutil"
)


func (fs *FileStore) DeleteChannel (channel string) bool{
	fs.RLock()
	delete(fs.channels, channel) //delete from in memory data structure

	//need to delete the file too..
	if (fs.Name() == TypeFile) {
		Noticef(fs.rootDir+"/"+channel)
		err := os.RemoveAll(fs.rootDir+"/"+channel)
		spew.Dump(err);
	}

	//spew.Dump(gs.channels[channel])
	fs.RUnlock()
	Noticef("deleted!")
	return true
}

func (ms *MemoryStore) DeleteChannel (channel string) bool{

	ms.RLock()
	delete(ms.channels, channel) //delete from in memory data structure
	ms.RUnlock()

	return true
}


func ReadFileStore(rootDir string) (*FileStore, *RecoveredState, error) {
	fs := &FileStore{
		rootDir: rootDir,
		opts:    DefaultFileStoreOptions,
	}

	fs.init(TypeFile, nil)

	//for _, opt := range options {
	//	if err := opt(&fs.opts); err != nil {
	//		return nil, nil, err
	//	}
	//}

	// Convert the compact interval in time.Duration
	fs.compactItvl = time.Duration(fs.opts.CompactInterval) * time.Second
	// Create the table using polynomial in options
	if fs.opts.CRCPolynomial == int64(crc32.IEEE) {
		fs.crcTable = crc32.IEEETable
	} else {
		fs.crcTable = crc32.MakeTable(uint32(fs.opts.CRCPolynomial))
	}

	if err := os.MkdirAll(rootDir, os.ModeDir+os.ModePerm); err != nil && !os.IsExist(err) {
		return nil, nil, fmt.Errorf("unable to create the root directory [%s]: %v", rootDir, err)
	}

	var err error
	//var recoveredState *RecoveredState
	var serverInfo *spb.ServerInfo
	var recoveredClients []*Client
	var recoveredSubs = make(RecoveredSubscriptions)
	var channels []os.FileInfo
	var msgStore *FileMsgStore
	var subStore *FileSubStore

	// Ensure store is closed in case of return with error
	defer func() {
		if err != nil {
			fs.Close()
		}
	}()

	// Open/Create the server file (note that this file must not be opened,
	// in APPEND mode to allow truncate to work).
	fileName := filepath.Join(fs.rootDir, serverFileName)
	fs.serverFile, err = openFile(fileName, os.O_RDWR, os.O_CREATE)
	if err != nil {
		return nil, nil, err
	}

	// Open/Create the client file.
	fileName = filepath.Join(fs.rootDir, clientsFileName)
	fs.clientsFile, err = openFile(fileName)
	if err != nil {
		return nil, nil, err
	}

	// Recover the server file.
	serverInfo, err = fs.recoverServerInfo()
	if err != nil {
		return nil, nil, err
	}

	err = os.MkdirAll("nats-streaming-server-info", 0774)
	outFile, err := os.Create("nats-streaming-server-info/server-info")
	if err != nil {
		panic(err)
	}

	outFile.WriteString("Server Information\n" + spew.Sdump(serverInfo) + "\n");

	// If the server file is empty, then we are done
	if serverInfo == nil {
		// We return the file store instance, but no recovered state.
		return fs, nil, nil
	}

	// Recover the clients file
	recoveredClients, err = fs.recoverClients()

	if err != nil {
		return nil, nil, err
	}

	outFile.WriteString("Connected Clients\n");
	outFile.WriteString(spew.Sdump(recoveredClients))

	outFile.Close();

	// Get the channels (there are subdirectories of rootDir)
	channels, err = ioutil.ReadDir(rootDir)
	if err != nil {
		return nil, nil, err
	}

	outFile, err = os.Create("nats-streaming-server-info/subject-list")
	if err != nil {
		panic(err)
	}

	for _, c := range channels {
		// Channels are directories. Ignore simple files
		if c.IsDir() {
			outFile.WriteString(spew.Sdump(c.Name()))
		}
	}

	outFile.Close()
	//spew.Dump(channels);

	// Go through the list
	for _, c := range channels {
		// Channels are directories. Ignore simple files
		if !c.IsDir() {
			//fmt.Println(c.Name());
			continue
		}
		channel := c.Name()
		channelDirName := filepath.Join(rootDir, channel)

		// Recover messages for this channel
		msgStore, err = fs.newFileMsgStore(channelDirName, channel, true)
		if err != nil {
			break
		}
		subStore, err = fs.newFileSubStore(channelDirName, channel, true)
		if err != nil {
			msgStore.Close()
			break
		}

		// For this channel, construct an array of RecoveredSubState
		rssArray := make([]*RecoveredSubState, 0, len(subStore.subs))

		// Fill that array with what we got from newFileSubStore.
		for _, sub := range subStore.subs {
			// The server is making a copy of rss.Sub, still it is not
			// a good idea to return a pointer to an object that belong
			// to the store. So make a copy and return the pointer to
			// that copy.
			csub := *sub.sub
			rss := &RecoveredSubState{
				Sub:     &csub,
				Pending: make(PendingAcks),
			}
			// If we recovered any seqno...
			if len(sub.seqnos) > 0 {
				// Lookup messages, and if we find those, update the
				// Pending map.
				for seq := range sub.seqnos {
					// Access directly 'msgs' here. If we have a
					// different implementation where we don't
					// keep messages around, we would still have
					// a cache of messages per channel that will
					// then be cleared after this loop when we
					// are done restoring the subscriptions.
					if m := msgStore.msgs[seq]; m != nil {
						rss.Pending[seq] = m
					}
				}
			}
			// Add to the array of recovered subscriptions
			rssArray = append(rssArray, rss)
		}

		// This is the recovered subscription state for this channel
		recoveredSubs[channel] = rssArray

		fs.channels[channel] = &ChannelStore{
			Subs: subStore,
			Msgs: msgStore,
		}
		err = os.MkdirAll("nats-streaming-server-info/"+channel, 0774)

		outFile, err = os.Create("nats-streaming-server-info/" + channel + "/msgs-list")
		if err != nil {
			panic(err)
		}
		outFile.WriteString(spew.Sdump(msgStore.genericMsgStore))
		outFile.Close()

		outFile, err = os.Create("nats-streaming-server-info/" + channel + "/subscription-list")
		if err != nil {
			panic(err)
		}
		outFile.WriteString(spew.Sdump(subStore.subs))
		outFile.Close()

		outFile, err = os.Create("nats-streaming-server-info/" + channel + "/recovered-subscription-list")
		if err != nil {
			panic(err)
		}
		outFile.WriteString(spew.Sdump(recoveredSubs[channel]))
		outFile.Close()

		//spew.Dump(recoveredSubs);

	}



	return nil, nil, nil;
}