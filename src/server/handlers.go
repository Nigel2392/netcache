package server

import (
	"net"
	"strconv"
	"strings"

	"github.com/Nigel2392/netcache/src/protocols"
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
	if s.logger != nil {
		s.logger.Debug("getting key")
	}
	var value, ttl, err = s.Cache.Get(message.Key)
	if err != nil {
		return err
	}
	message.Value = value
	message.TTL = ttl
	if s.logger != nil {
		s.logger.Debug("sending response")
	}
	_, err = message.WriteTo(c)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleSet(c net.Conn, message *protocols.Message) error {
	if s.logger != nil {
		s.logger.Debug("setting key")
	}
	var _, err = s.Cache.Set(message.Key, message.Value, message.TTL)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleDelete(c net.Conn, message *protocols.Message) error {
	if s.logger != nil {
		s.logger.Debug("deleting key")
	}
	var _, err = s.Cache.Delete(message.Key)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleClear(c net.Conn) error {
	if s.logger != nil {
		s.logger.Debug("clearing cache")
	}
	var err = s.Cache.Clear()
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleHas(c net.Conn, message *protocols.Message) error {
	if s.logger != nil {
		s.logger.Debug("checking if key exists")
	}
	var _, has = s.Cache.Has(message.Key)
	message.Value = []byte(strconv.FormatBool(has))
	if s.logger != nil {
		s.logger.Debugf("sending has: %v\n", string(message.Value))
	}
	var _, err = message.WriteTo(c)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handleKeys(c net.Conn) error {
	if s.logger != nil {
		s.logger.Debug("fetching keys")
	}
	var keys = s.Cache.Keys()
	var message = &protocols.Message{
		Type:  protocols.TypeKEYS,
		Value: []byte(strings.Join(keys, ",")),
	}
	if s.logger != nil {
		s.logger.Debugf("sending keys: %v\n", string(message.Value))
	}
	var _, err = message.WriteTo(c)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheServer) handlePing(c net.Conn) error {
	if s.logger != nil {
		s.logger.Debug("pinging")
	}
	var message = &protocols.Message{
		Type: protocols.TypePONG,
	}
	if s.logger != nil {
		s.logger.Debug("sending pong")
	}
	var _, err = message.WriteTo(c)
	if err != nil {
		return err
	}
	return nil
}
