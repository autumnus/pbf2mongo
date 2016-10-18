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
