package main

import (
	"io"

	"github.com/jackpal/bencode-go"
)

type FileInfo struct {
	Name        string "name"
	Length      int64  "length"
	Pieces      string "pieces"
	PieceLength int    "piece length"
}

type TorrentInfo struct {
	Announce     string     "announce"
	AnnounceList [][]string "announce-list"
	Info         FileInfo   "info"
	HttpSeeds    []string   "httpseeds"
}

func generateTorrent(writer io.Writer, url string) error {
	info := FileInfo{
		Name:        "file.mp3",
		Length:      1024,
		Pieces:      "01234567890123456789",
		PieceLength: 1024,
	}
	announceList := [][]string{[]string{url}}
	seeds := []string{url}
	torrent := TorrentInfo{Info: info, Announce: url, HttpSeeds: seeds, AnnounceList: announceList}
	err := bencode.Marshal(writer, torrent)
	return err
}
