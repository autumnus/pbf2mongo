package main

import (
	"github.com/qedus/osmpbf"
	"github.com/cheggaaa/pb"
	"labix.org/v2/mgo"

	"flag"
	"log"
	"os"
	"runtime"
	"time"
	"io"
	"fmt"
)

var fileName = flag.String("f", "", "OSM XML file to import.")
var mongoAddr = flag.String("s", "127.0.0.1:27017", "MongoDB server to import to.")
var mongoDbName = flag.String("db", "osm", "MongoDB database name (defaults to 'osm'.")
var mongoSession *mgo.Session
var insertChan chan interface{}

//   <node id="2031042144" version="1" timestamp="2012-11-24T23:19:36Z" uid="560392" user="HostedDinner" changeset="14021355" lat="25.5805168" lon="-80.3562449"/>
type OsmNode struct {
	Id  int64 `bson:"_id" xml:"id,attr"`
	Type string `bson:"type"`
	Loc struct {
		    Type        string    `bson:"type"`
		    Coordinates []float64 `bson:"coordinates"`
	    } `bson:"loc"`
	Version   int32     `bson:"version"       xml:"version,attr"`
	Ts        time.Time `bson:"timestamp"        xml:"timestamp,attr"`
	Uid       int64     `bson:"uid"       xml:"uid,attr"`
	User      string    `bson:"user"      xml:"user,attr"`
	ChangeSet int64     `bson:"changeset" xml:"changeset,attr"`
	Lat       float64   `bson:"-"         xml:"lat,attr"`
	Lng       float64   `bson:"-"         xml:"lon,attr"`
	Tags      map[string]string `bson:"tags"`
	//RTags     []struct {
	//	Key   string `bson:"-" xml:"k,attr"`
	//	Value string `bson:"-" xml:"v,attr"`
	//} `bson:"-"        xml:"tag"`
}

/*
  <way id="11137619" version="2" timestamp="2013-02-05T23:54:16Z" uid="451693" user="bot-mode" changeset="14928391">
    <nd ref="99193738"/>
    <nd ref="99193742"/>
    <nd ref="99193745"/>
    <nd ref="99193748"/>
    <nd ref="99193750"/>
    <nd ref="99147506"/>
    <tag k="highway" v="residential"/>
    <tag k="name" v="Southwest 148th Avenue Court"/>
    <tag k="tiger:cfcc" v="A41"/>
    <tag k="tiger:county" v="Miami-Dade, FL"/>
    <tag k="tiger:name_base" v="148th Avenue"/>
    <tag k="tiger:name_direction_prefix" v="SW"/>
    <tag k="tiger:name_type" v="Ct"/>
    <tag k="tiger:reviewed" v="no"/>
    <tag k="tiger:zip_left" v="33185"/>
    <tag k="tiger:zip_right" v="33185"/>
  </way>*/
type OsmWay struct {
	Id  int64 `bson:"_id" xml:"id,attr"`
	Type string `bson:"type"`
	Center struct {
		    Type        string      `bson:"type"`
		    Center      []float64   `bson:"coordinates"`
	    } `bson:"loc"`
	Version   int32             `bson:"version"       xml:"version,attr"`
	Ts        time.Time         `bson:"timestamp"        xml:"timestamp,attr"`
	Uid       int64             `bson:"uid"       xml:"uid,attr"`
	User      string            `bson:"user"      xml:"user,attr"`
	ChangeSet int64             `bson:"changeset" xml:"changeset,attr"`
	Tags      map[string]string `bson:"tags"`
	//RTags     []struct {
	//	Key   string `bson:"-" xml:"k,attr"`
	//	Value string `bson:"-" xml:"v,attr"`
	//} `bson:"-" xml:"tag"`
	NodeIDs []int64
	//Nds []struct {
	//	Id int64 `bson:"-" xml:"ref,attr"`
	//} `bson:"-"         xml:"nd"`
	//Loc struct {
	//	    Type        string      `bson:"type"`
	//	    Nodes [][]float64 `bson:"nodes"`
	//    } `bson:"nodes"`
}

func main() {
	var err error

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	mongoSession, err = mgo.Dial(*mongoAddr)
	if err != nil {
		log.Fatalln("Can't connect to MongoDB: " + err.Error())
	}

	index := mgo.Index{
		Key: []string{"$2dsphere:loc"},
	}

	log.Println("Preparing database & collections...")

	mongoSession.DB(*mongoDbName+"_nodes").DropDatabase()
	mongoSession.DB(*mongoDbName+"_nodes").C("data").EnsureIndex(index)

	mongoSession.DB(*mongoDbName+"_ways").DropDatabase()
	mongoSession.DB(*mongoDbName+"_ways").C("data").EnsureIndex(index)

	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatalln("Can't open OSM file: " + err.Error())
	}

	insertChan = make(chan interface{}, 100)
	go goInsert()
	go goInsert()
	go goInsert()
	go goInsert()

	decoder2 := osmpbf.NewDecoder(file)
	err = decoder2.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	if err != nil {
		log.Fatal(err)
	}
	var nc, wc, rc uint64

	log.Println("Processing Nodes and Ways...")
	stat, _ := file.Stat()
	bar := pb.New(int(stat.Size() / 1024)).SetUnits(pb.U_NO)
	bar.Start()


	// TODO Ways will fail, if contained nodes are not in DB!! So possibly make one loop foreach type!!
	for {
		if v, err := decoder2.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				// Process Node v.
				var n OsmNode
				n.Id = v.ID
				n.Type = "node"
				n.Loc.Type = "Point"
				n.Loc.Coordinates = []float64{v.Lon, v.Lat}
				n.Lat = v.Lat
				n.Lng = v.Lon
				n.Tags = make(map[string]string)
				for key, val := range v.Tags {
					n.Tags[key]=val
				}
				n.User = v.Info.User
				n.Ts = v.Info.Timestamp
				n.Version = v.Info.Version
				n.ChangeSet = v.Info.Changeset
				n.Uid = v.Info.Uid

				insertChan <- n

				nc++
			case *osmpbf.Way:
				// Process Way v.
				var w OsmWay
				w.Id = v.ID
				w.Type = "way"
				//w.Center.Type = "Point"
				//w.Center.Center = []float64{0, 0}
				w.Tags = v.Tags
				w.User = v.Info.User
				w.Ts = v.Info.Timestamp
				w.Version = v.Info.Version
				w.ChangeSet = v.Info.Changeset
				w.Uid = v.Info.Uid
				w.NodeIDs = v.NodeIDs

				insertChan <- w

				wc++
			case *osmpbf.Relation:
				// Process Relation v.
				rc++
			default:
				log.Fatalf("unknown type %T\n", v)
			}
		}
		ofs, _ := file.Seek(0, 1)
		bar.Set((int)(ofs / 1024))
	}
	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)

	file.Seek(0, 0)

	return
}

func goInsert() {
	sess := mongoSession.Clone()

	for {
		select {
		case i := <-insertChan:
			switch o := i.(type) {
			case OsmNode:
				o.Loc.Type = "Point"
				o.Loc.Coordinates = []float64{o.Lng, o.Lat}

				err := sess.DB(*mongoDbName+"_nodes").C("data").Insert(o)
				if err != nil {
					log.Println(err.Error())
				}

			case OsmWay:
				var n OsmNode
				//o.Loc.Type = "LineString"
				//o.Loc.Nodes = make([][]float64,0,len(o.Nds))

				var lat_i0 = float64(0)
				var lat_i1 = float64(0)
				var sum_x = float64(0)
				var lng_i0 = float64(0)
				var lng_i1 = float64(0)
				var sum_y = float64(0)
				var A = float64(0)
				var i = 0

				for _, nid := range o.NodeIDs {
					//if  sess.DB(*mongoDbName+"_nodes").C("data").Find(bson.D{{"id", nid}}).One(&n)==nil {
					if  sess.DB(*mongoDbName+"_nodes").C("data").FindId(nid).One(&n)==nil {
						//o.Loc.Nodes = append(o.Loc.Nodes,[]float64{n.Loc.Coordinates[0],n.Loc.Coordinates[1]})
						if i == 0 {
							lng_i0 = n.Loc.Coordinates[0]
							lat_i0 = n.Loc.Coordinates[1]
						} else {
							lng_i1 = n.Loc.Coordinates[0]
							lat_i1 = n.Loc.Coordinates[1]

							sum_x += (lng_i0 + lng_i1) * (lng_i0 * lat_i1 - lng_i1 * lat_i0)
							sum_y += (lat_i0 + lat_i1) * (lng_i0 * lat_i1 - lng_i1 * lat_i0)

							A += (lng_i0*lat_i1 - lng_i1*lat_i0)

							lng_i0 = lng_i1
							lat_i0 = lat_i1
						}
					}

					i = 1
				}
				A = A / 2;
				var x,y float64
				if (6*A) *sum_x == 0 {
					y = n.Loc.Coordinates[0]
					x = n.Loc.Coordinates[1]
				} else {
					x = 1 / (6 * A) * sum_x
					y = 1 / (6 * A) * sum_y
				}
				o.Center.Type = "Point"
				o.Center.Center = []float64{x, y}

				err :=  sess.DB(*mongoDbName+"_ways").C("data").Insert(o)
				if err != nil {
					log.Println("")
					log.Println(err.Error())
					fmt.Sprintf("%#v", o)
					log.Println("")
				}
			}
		}
	}
}