package main

import (
	"flag"
	"fmt"
	"github.com/kid0m4n/gorays/vector"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
)

var art = []string{
	"                   ",
	"    1111           ",
	"   1    1          ",
	"  1           11   ",
	"  1          1  1  ",
	"  1     11  1    1 ",
	"  1      1  1    1 ",
	"   1     1   1  1  ",
	"    11111     11   ",
}

var G []int = makeBV()

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	width      = flag.Int("width", 512, "width of the rendered image")
	height     = flag.Int("height", 512, "height of the rendered image")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	fmt.Printf("P6 %v %v 255 ", *width, *height)

	g := vector.Vector{-5.5, -16, 0}.Normalize()
	a := vector.Vector{0, 0, 1}.CrossProduct(g).Normalize().Scale(0.002)
	b := g.CrossProduct(a).Normalize().Scale(0.002)
	c := a.Add(b).Scale(-256).Add(g)

	for y := (*height - 1); y >= 0; y-- {
		for x := (*width - 1); x >= 0; x-- {
			p := vector.Vector{13, 13, 13}

			for i := 0; i < 64; i++ {
				t := a.Scale(rand.Float64() - 0.5).Scale(99).Add(b.Scale(rand.Float64() - 0.5).Scale(99))
				orig := vector.Vector{17, 16, 8}.Add(t)
				dir := t.Scale(-1).Add(a.Scale(rand.Float64() + float64(x)).Add(b.Scale(float64(y) + rand.Float64())).Add(c).Scale(16)).Normalize()
				p = sampler(orig, dir).Scale(3.5).Add(p)
			}

			if n, err := os.Stdout.Write([]byte{byte(p.X), byte(p.Y), byte(p.Z)}); n != 3 || err != nil {
				panic(err)
			}
		}
	}
}

func sampler(orig, dir vector.Vector) vector.Vector {
	st, dist, bounce := tracer(orig, dir)

	if st == missUpward {
		return vector.Vector{0.7, 0.6, 1}.Scale(math.Pow(1-dir.Z, 4))
	}

	h := orig.Add(dir.Scale(dist))
	l := vector.Vector{9 + rand.Float64(), 9 + rand.Float64(), 16}.Add(h.Scale(-1)).Normalize()
	r := dir.Add(bounce.Scale(bounce.DotProduct(dir.Scale(-2))))

	b := l.DotProduct(bounce)

	if b < 0 {
		b = 0
	} else {
		var st status
		if st, dist, bounce = tracer(h, l); st != missUpward {
			b = 0
		}
	}

	var sf float64
	if b > 0 {
		sf = 1.0
	}

	p := math.Pow(l.DotProduct(r.Scale(sf)), 99)

	if st == missDownward {
		h = h.Scale(0.2)
		fc := vector.Vector{3, 3, 3}
		if int(math.Ceil(h.X)+math.Ceil(h.Y))&1 == 1 {
			fc = vector.Vector{3, 1, 1}
		}
		return fc.Scale(b*0.2 + 0.1)
	}

	return vector.Vector{p, p, p}.Add(sampler(h, r).Scale(0.5))
}

type status int

const (
	missUpward = iota
	missDownward
	hit
)

func tracer(orig, dir vector.Vector) (st status, dist float64, bounce vector.Vector) {
	dist = 1e9
	st = missUpward
	p := -orig.Z / dir.Z
	if 0.01 < p {
		dist = p
		bounce = vector.Vector{0, 0, 1}
		st = missDownward
	}

	for k := 18; k >= 0; k-- {
		for j, g := range G {
			if g&(1<<uint(k)) != 0 {
				p := orig.Add(vector.Vector{float64(-k), 3, float64(-j - 4)})
				b := p.DotProduct(dir)
				c := p.DotProduct(p) - 1
				q := b*b - c

				if q > 0 {
					s := -b - math.Sqrt(q)

					if s < dist && s > 0.01 {
						dist = s
						bounce = p.Add(dir.Scale(dist)).Normalize()
						st = hit
					}
				}
			}
		}
	}

	return
}

func makeBV() []int {
	bv := make([]int, 9)
	for i, s := range art {
		sum := 0
		for j, c := range s {
			if string(c) != " " {
				sum += 1 << uint(18-j)
			}
		}

		bv[8-i] = sum
	}

	return bv
}
