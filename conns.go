package moetcpserver

var Conns *SyncMap = SyncmapNew(delete_map, set_map)
