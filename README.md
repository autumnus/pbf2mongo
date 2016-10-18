# pbf2mongo

This project is a "first time GO project". It is more or less a combination of two github projects [bocajim/goosm](https://github.com/bocajim/goosm) and [qedus/osmpbf](https://github.com/qedus/osmpbf) to push OSM PBF files into a MongoDB

The usage is like goosm
### Running

* -f \<osm.pbf file to read\>
* -s \<mongo server:port to connect to\> (127.0.0.1:27017 default)
* -db \<name of mongodb database\> (osm default)

### Examples

pbf2mongo -f miami.osm.pbf

pbf2mongo -f miami.osm.pbf -db foo

pbf2mongo -f miami.osm.pbf -s 127.0.0.1:27017

### DB Structure

#### Nodes
```
{ 
    "_id" : NumberLong(24084961), 
    "type" : "node", 
    "loc" : {
        "type" : "Point", 
        "coordinates" : [
            13.111418, 
            -12.923375
        ]
    }, 
    "version" : NumberInt(11), 
    "timestamp" : ISODate("2007-11-24T16:00:38.000+0000"), 
    "uid" : NumberLong(73014447592), 
    "user" : "Firefishy", 
    "changeset" : NumberLong(606620), 
    "tags" : {
        "source" : "PGS", 
        "created_by" : "almien_coastlines"
    }
}
```

#### Ways
For ways a centroid (loc.coordinates) is computed from the coordinates of the referenced nodes.
```
{ 
    "_id" : NumberLong(4098264), 
    "type" : "way", 
    "loc" : {
        "type" : "Point", 
        "coordinates" : [
            11.757986170094167, 
            -17.25621459658663
        ]
    }, 
    "version" : NumberInt(4), 
    "timestamp" : ISODate("2013-01-27T19:18:15.000+0000"), 
    "uid" : NumberLong(212111), 
    "user" : "okilimu", 
    "changeset" : NumberLong(14812629), 
    "tags" : {
        "source" : "Bing"
    }, 
    "nodeids" : [
        NumberLong(21936815), 
        NumberLong(21936838), 
        NumberLong(21936834), 
        NumberLong(21936829), 
        NumberLong(21936826), 
        NumberLong(21936824), 
        NumberLong(21936821), 
        NumberLong(2131273016), 
        NumberLong(21936820), 
        NumberLong(21936815)
    ]
}
```

#### Relations
Relations are not (yet) handeled
