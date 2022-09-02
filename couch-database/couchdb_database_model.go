package couch_database

/*
{
  "db_name": "quotes",
  "purge_seq": "0-g1AAAABPeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCeexAEmGBiD1HwiyEhlwqEtkSKqHKMgCAIT2GV4",
  "update_seq": "5814-g1AAAABVeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCeexAEmGBiD1HwiykhgYeNRwKE1kSKqHquGakAUAAI0aLA",
  "sizes": {
    "file": 32686514,
    "external": 29415443,
    "active": 19217279
  },
  "props": {
    "partitioned": true
  },
  "doc_del_count": 0,
  "doc_count": 5758,
  "disk_format_version": 8,
  "compact_running": false,
  "cluster": {
    "q": 2,
    "n": 1,
    "w": 1,
    "r": 1
  },
  "instance_start_time": "0"
}
*/

type ClusterInfo struct {
	Shards      int `json:"q"`
	Replicas    int `json:"n"`
	WriteQuorum int `json:"w"`
	ReadQuorum  int `json:"r"`
}

type DatabaseSizes struct {
	File     int64 `json:"file"`
	External int64 `json:"external"`
	Active   int64 `json:"active"`
}

type DatabaseProperty struct {
	Partitioned bool `json:"partitioned"`
}

type CouchDatabaseInfo struct {
	DatabaseName        string           `json:"db_name"`
	PurgeSequence       string           `json:"purge_seq"`
	UpdateSequence      string           `json:"update_seq"`
	Sizes               DatabaseSizes    `json:"sizes"`
	Properties          DatabaseProperty `json:"props"`
	DocumentDeleteCount int64            `json:"doc_del_count"`
	DocumentCount       int64            `json:"doc_count"`
	Cluster             ClusterInfo      `json:"cluster"`
	DiskFormatVersion   int              `json:"disk_format_version"`
	CompactRunning      bool             `json:"compact_running"`
}
