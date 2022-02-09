package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/require"
)

func TestNewWatcher(t *testing.T) {
	w, err := New(func(event *WatcherEvent) error {
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	require.NotNil(t, w)
}

func TestWatcherAddRemoveExist(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	require.NoError(t, err)
	file2 := testutil.TempFile(t, "temp_config"+randomString(12))
	_, err = file2.WriteString("test config 2")
	require.NoError(t, err)
	err = file2.Sync()
	require.NoError(t, err)
	err = file2.Close()
	require.NoError(t, err)

	require.NoError(t, err)
	file3 := testutil.TempFile(t, "temp_config"+randomString(12))
	_, err = file3.WriteString("test config 2")
	require.NoError(t, err)
	err = file3.Sync()
	require.NoError(t, err)
	err = file3.Close()
	require.NoError(t, err)

	err = w.Add(file.Name())
	require.NoError(t, err)

	h, ok := w.configFiles[file.Name()]
	require.True(t, ok)
	require.NotEqual(t, 0, h.iNode)

	require.NoError(t, err)
	w.Start()
	err = os.Rename(file2.Name(), file.Name())
	require.NoError(t, err)
	require.NoError(t, assertEvent(file.Name(), watcherCh))

	// wait for file to be added back
	time.Sleep(w.reconcileTimeout + 50*time.Millisecond)
	err = w.Remove(file.Name())
	require.NoError(t, err)
	_, ok = w.configFiles[file.Name()]
	require.False(t, ok)

	err = os.Rename(file3.Name(), file.Name())
	require.NoError(t, err)
	require.Error(t, assertEvent(file.Name(), watcherCh), "timedout waiting for event")
}

func TestWatcherAddNotExist(t *testing.T) {
	w, err := New(func(event *WatcherEvent) error {
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	filename := file.Name() + randomString(16)
	err = w.Add(filename)
	require.True(t, os.IsNotExist(err))
	_, ok := w.configFiles[filename]
	require.False(t, ok)
}

func TestEventWatcherWrite(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)

	err = w.Add(file.Name())
	require.NoError(t, err)
	w.Start()
	_, err = file.WriteString("test config 2")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)
	require.Error(t, assertEvent(file.Name(), watcherCh), "timedout waiting for event")
}

func TestEventWatcherRead(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)

	err = w.Add(file.Name())
	require.NoError(t, err)
	w.Start()
	_, err = os.ReadFile(file.Name())
	require.NoError(t, err)
	require.Error(t, assertEvent(file.Name(), watcherCh), "timedout waiting for event")
}

func TestEventWatcherChmod(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)

	err = w.Add(file.Name())
	require.NoError(t, err)
	w.Start()
	file.Chmod(0777)
	require.NoError(t, err)
	require.Error(t, assertEvent(file.Name(), watcherCh), "timedout waiting for event")
}

func TestEventWatcherRemoveCreate(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	err = w.Add(file.Name())
	require.NoError(t, err)
	w.reconcileTimeout = 20 * time.Millisecond
	w.Start()
	err = os.Remove(file.Name())
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
	recreated, err := os.Create(file.Name())
	require.NoError(t, err)
	_, err = recreated.WriteString("config 2")
	require.NoError(t, err)
	err = recreated.Sync()
	require.NoError(t, err)
	// this an event coming from the reconcile loop
	require.NoError(t, assertEvent(file.Name(), watcherCh))
	iNode, err := w.getFileId(recreated.Name())
	require.NoError(t, err)
	require.Equal(t, iNode, w.configFiles[recreated.Name()].iNode)
}

func TestEventWatcherMove(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	w.reconcileTimeout = 20 * time.Millisecond
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		file2 := testutil.TempFile(t, "temp_config"+randomString(12))
		_, err = file2.WriteString("test config 2")
		require.NoError(t, err)
		err = file2.Sync()
		require.NoError(t, err)
		err = file2.Close()
		require.NoError(t, err)

		err = w.Add(file.Name())

		require.NoError(t, err)
		w.Start()
		err = os.Rename(file2.Name(), file.Name())
		require.NoError(t, err)
		require.NoError(t, assertEvent(file.Name(), watcherCh))
		time.Sleep(w.reconcileTimeout + 10*time.Millisecond)
		iNode, err := w.getFileId(file.Name())
		require.NoError(t, err)
		require.Equal(t, iNode, w.configFiles[file.Name()].iNode)

	}
}

func TestEventReconcileMove(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)

	file2 := testutil.TempFile(t, "temp_config"+randomString(12))
	_, err = file2.WriteString("test config 2")
	require.NoError(t, err)
	err = file2.Sync()
	require.NoError(t, err)
	err = file2.Close()
	require.NoError(t, err)

	err = w.Add(file.Name())
	require.NoError(t, err)
	w.reconcileTimeout = 20 * time.Millisecond
	// remove the file from the internal watcher to only trigger the reconcile
	err = w.watcher.Remove(file.Name())
	require.NoError(t, err)
	w.Start()
	err = os.Rename(file2.Name(), file.Name())
	require.NoError(t, err)
	require.NoError(t, assertEvent(file.Name(), watcherCh))
	iNode, err := w.getFileId(file.Name())
	require.NoError(t, err)
	require.Equal(t, iNode, w.configFiles[file.Name()].iNode)
}

func assertEvent(name string, watcherCh chan *WatcherEvent) error {
	timeout := time.After(1000 * time.Millisecond)
	select {
	case ev := <-watcherCh:
		if ev.Filename != name {
			return fmt.Errorf("filename do not match")
		}
		return nil
	case <-timeout:
		return fmt.Errorf("timedout waiting for event")
	}
}