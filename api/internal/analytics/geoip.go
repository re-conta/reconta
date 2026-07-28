package analytics

import (
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// GeoIP resolve país/região/cidade/coordenadas a partir de um IP, usando o
// banco GeoLite2-City da MaxMind. Fica desabilitado (Lookup retorna zero
// values) quando nenhum caminho de banco é informado ou o arquivo não existe
// — geolocalização é um extra, nunca deve impedir o registro da visita.
type GeoIP struct {
	path   string
	reader atomic.Pointer[geoip2.Reader]
}

// NewGeoIP abre o banco em path (se informado) e agenda recarregamentos
// periódicos, para acompanhar as atualizações que o geoipupdate grava no
// disco da VPS sem precisar reiniciar o processo.
func NewGeoIP(path string) *GeoIP {
	g := &GeoIP{path: path}
	if path == "" {
		log.Print("GEOIP_DB_PATH não definido: geolocalização de visitas desabilitada")
		return g
	}

	g.reload()
	if g.reader.Load() == nil {
		log.Printf("GeoIP: banco não encontrado em %s, geolocalização desabilitada até o próximo recarregamento", path)
	}

	go func() {
		for range time.Tick(time.Hour) {
			g.reload()
		}
	}()

	return g
}

func (g *GeoIP) reload() {
	if g.path == "" {
		return
	}
	reader, err := geoip2.Open(g.path)
	if err != nil {
		return
	}
	if old := g.reader.Swap(reader); old != nil {
		old.Close()
	}
}

type geoResult struct {
	Country   string
	Region    string
	City      string
	Latitude  float64
	Longitude float64
}

// Lookup retorna a localização aproximada do IP, ou geoResult zero-value se a
// geolocalização estiver desabilitada, o IP for inválido/privado ou não
// houver correspondência no banco.
func (g *GeoIP) Lookup(ip string) geoResult {
	reader := g.reader.Load()
	if reader == nil || ip == "" {
		return geoResult{}
	}

	record, err := reader.City(net.ParseIP(ip))
	if err != nil || record == nil {
		return geoResult{}
	}

	region := ""
	if len(record.Subdivisions) > 0 {
		region = record.Subdivisions[0].Names["en"]
	}

	return geoResult{
		Country:   record.Country.Names["en"],
		Region:    region,
		City:      record.City.Names["en"],
		Latitude:  record.Location.Latitude,
		Longitude: record.Location.Longitude,
	}
}
