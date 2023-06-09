package symtab

import (
	"debug/elf"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func Test(t *testing.T) {
	t.Skip()
	fs := []string{
		"/Users/korniltsev/Downloads/dump/3026076/bin/external-secrets",
		"/Users/korniltsev/Downloads/dump/1952222/bin/bash",
		"/Users/korniltsev/Downloads/dump/1720591/usr/local/bin/runsvdir",
		"/Users/korniltsev/Downloads/dump/408549/usr/bin/promtail",
		"/Users/korniltsev/Downloads/dump/5342/ip-masq-agent",
		"/Users/korniltsev/Downloads/dump/2271859/usr/bin/grafana-agent",
		"/Users/korniltsev/Downloads/dump/115664/usr/local/bin/node",
		"/Users/korniltsev/Downloads/dump/2554161/usr/bin/cloudwatch-exporter",
		"/Users/korniltsev/Downloads/dump/6785/usr/bin/cloud-backend-gateway",
		"/Users/korniltsev/Downloads/dump/131427/usr/lib/jvm/java-1.8-openjdk/jre/bin/java",
		"/Users/korniltsev/Downloads/dump/176439/bin/mysqld_exporter",
		"/Users/korniltsev/Downloads/dump/197795/bin/agent",
		"/Users/korniltsev/Downloads/dump/977178/home/ray/anaconda3/bin/python3.9",
		"/Users/korniltsev/Downloads/dump/524479/cloud_sql_proxy",
		"/Users/korniltsev/Downloads/dump/236969/app/conntrack-exporter",
		"/Users/korniltsev/Downloads/dump/96262/usr/bin/auth-proxy",
		"/Users/korniltsev/Downloads/dump/3658358/tempo",
		"/Users/korniltsev/Downloads/dump/3875540/usr/bin/grafana-agent",
		"/Users/korniltsev/Downloads/dump/187314/usr/sbin/collectd",
		"/Users/korniltsev/Downloads/dump/168113/usr/sbin/nginx",
		"/Users/korniltsev/Downloads/dump/119326/usr/sbin/collectd",
		"/Users/korniltsev/Downloads/dump/792731/usr/sbin/nginx",
		"/Users/korniltsev/Downloads/dump/5120/bin/node_exporter",
		"/Users/korniltsev/Downloads/dump/482093/usr/bin/grafana-agent",
		"/Users/korniltsev/Downloads/dump/861838/bin/mysqld_exporter",
		"/Users/korniltsev/Downloads/dump/44296/usr/sbin/collectd",
		"/Users/korniltsev/Downloads/dump/2618433/opt/bitnami/java/bin/java",
		"/Users/korniltsev/Downloads/dump/1567047/grl-exporter",
		"/Users/korniltsev/Downloads/dump/4085865/usr/local/bin/memcached",
		"/Users/korniltsev/Downloads/dump/1055834/usr/bin/promtail",
		"/Users/korniltsev/Downloads/dump/10386/bin/memcached_exporter",
		"/Users/korniltsev/Downloads/dump/470140/usr/bin/promtail",
		"/Users/korniltsev/Downloads/dump/114399/cloud_sql_proxy",
		"/Users/korniltsev/Downloads/dump/4132246/usr/bin/enterprise-metrics",
		"/Users/korniltsev/Downloads/dump/978887/usr/bin/cloud-backend-gateway",
		"/Users/korniltsev/Downloads/dump/436714/usr/bin/promtail",
		"/Users/korniltsev/Downloads/dump/3104561/usr/bin/loki-canary",
		"/Users/korniltsev/Downloads/dump/316471/app/cmd/cainjector/cainjector.runfiles/com_github_jetstack_cert_manager/cmd/cainjector/cainjector_/cainjector",
		"/Users/korniltsev/Downloads/dump/287987/usr/local/bin/crossplane-terraform-provider",
		"/Users/korniltsev/Downloads/dump/4085904/bin/memcached_exporter",
	}
	for _, f := range fs {

		inspect(t, f)
	}

}

func inspect(t *testing.T, f string) {
	fmt.Println()
	fmt.Println(f)
	ff, err2 := os.Open(f)
	if err2 != nil {
		panic(err2)
	}
	_ = ff
	e, err := elf.NewFile(ff)

	if err != nil {
		fmt.Println(err)
		return
	}
	var genuineSymbols []Sym
	symbols, _ := e.Symbols()
	dynSymbols, _ := e.DynamicSymbols()
	namesLength := 0
	count := 0
	for _, symbol := range symbols {
		if symbol.Value != 0 && symbol.Info&0xf == byte(elf.STT_FUNC) {
			namesLength += len(symbol.Name)
			count += 1
			genuineSymbols = append(genuineSymbols, Sym{
				Name:  symbol.Name,
				Start: symbol.Value,
			})
		}
	}
	for _, symbol := range dynSymbols {
		if symbol.Value != 0 && symbol.Info&0xf == byte(elf.STT_FUNC) {
			namesLength += len(symbol.Name)
			count += 1
			genuineSymbols = append(genuineSymbols, Sym{
				Name:  symbol.Name,
				Start: symbol.Value,
			})
		}
	}
	fmt.Printf("names len %d cnt %d\n", namesLength, count)

	me, err := NewMMapedElfFile(f)
	defer me.close()

	me.readSymbols()
	require.NoError(t, err)
	var mySymbols []Sym

	for i := range me.symbols {
		sym := &me.symbols[i]
		name, _ := me.symbolName(sym)
		mySymbols = append(mySymbols, Sym{
			Name:  name,
			Start: sym.Value,
		})
	}

	cmp := func(a, b Sym) bool {
		if a.Start == b.Start {
			return strings.Compare(a.Name, b.Name) < 0
		}
		return a.Start < b.Start
	}
	slices.SortFunc(genuineSymbols, cmp)
	slices.SortFunc(mySymbols, cmp)
	require.Equal(t, genuineSymbols, mySymbols)

}
