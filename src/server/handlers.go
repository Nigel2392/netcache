package server

import (
	"net"
	"strconv"
	"strings"

	"github.com/Nigel2392/netcache/src/internal/protocols"
)

func writeErrorMessage(c net.Conn, err error) error {
	var message = &protocols.Message{
		Type:  protocols.TypeERROR,
		Value: []byte(err.Error()),
	}

	_, err = message.WriteTo(c)
	return err
}

func (s *CacheServer) handleGet(c net.Conn, message *protocols.Message) error {
	var value, ttl, err = s.cache.Get(message.Key)
	if err != nil {
		return err
	}
	message.Value = value
	message.TTL = ttl
	_, err = message.WriteTo(c)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleSet(c net.Conn, message *protocols.Message) error {
	var _, err = s.cache.Set(message.Key, message.Value, message.TTL)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleDelete(c net.Conn, message *protocols.Message) error {
	var _, err = s.cache.Delete(message.Key)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleClear(c net.Conn) error {
	var err = s.cache.Clear()
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleHas(c net.Conn, message *protocols.Message) error {
	var _, has = s.cache.Has(message.Key)
	message.Value = []byte(strconv.FormatBool(has))
	var _, err = message.WriteTo(c)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleKeys(c net.Conn) error {
	var keys = s.cache.Keys()
	var message = &protocols.Message{
		Type:  protocols.TypeKEYS,
		Value: []byte(strings.Join(keys, ",")),
	}
	var _, err = message.WriteTo(c)
	if err != nil {
		return err
	}
	return nil
}
