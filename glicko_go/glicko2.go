package glicko_go

import (
	"errors"
	"math"
)

const TAU = 0.5
const FACTOR = 173.7178
const EPSILON = 0.000001

const (
	WIN  float64 = 1
	TIE  float64 = 0.5
	LOSE float64 = 0
)

type Glicko1 struct {
	Rating float64
	Sigma  float64 // std.dev
	Rd     float64
}

type Glicko2 struct {
	Mu    float64 // mean
	Sigma float64 // std. dev
	Phi   float64 // rating deviation
}

func Glicko2From1(g1 Glicko1) Glicko2 {
	return Glicko2{
		Mu:    (g1.Rating - 1500) / FACTOR,
		Sigma: g1.Sigma,
		Phi:   g1.Rd / FACTOR,
	}
}

func Glicko1From2(glicko2 Glicko2) Glicko1 {
	return Glicko1{
		Rating: glicko2.Mu*FACTOR + 1500,
		Sigma:  glicko2.Sigma,
		Rd:     glicko2.Phi * FACTOR,
	}
}

// E
func e(mu float64, muJ float64, phiJ float64) float64 {
	return 1 / (1 + math.Exp(-g(phiJ)*(mu-muJ)))
}

func g(phi float64) float64 {
	return 1 / math.Sqrt(1+3*math.Pow(phi, 2)/math.Pow(math.Pi, 2))
}

func computeV(gCur Glicko2, gOpponents []Glicko2) float64 {
	var sum float64 = 0
	for _, gOp := range gOpponents {
		sum += math.Pow(g(gOp.Phi), 2) * e(gCur.Mu, gOp.Mu, gOp.Phi) * (1 - e(gCur.Mu, gOp.Mu, gOp.Phi))
	}
	return 1 / sum
}

func computeDelta(gCur Glicko2, gOpponents []Glicko2, scores []float64) float64 {
	var sum float64 = 0
	for j, gOp := range gOpponents {
		sum += g(gOp.Phi) * (scores[j] - e(gCur.Mu, gOp.Mu, gOp.Phi)) // S instead of 1
	}
	return computeV(gCur, gOpponents) * sum
}

func sigmaByIllinois(gCur Glicko2, delta float64, v float64) float64 {
	a := math.Log(math.Pow(gCur.Sigma, 2))
	f := func(x float64) float64 {
		t1 := math.Exp(x) * (math.Pow(delta, 2) - math.Pow(gCur.Phi, 2) - v - math.Exp(x)) / 2 * math.Pow(math.Pow(gCur.Phi, 2)+v+math.Exp(x), 2)
		t2 := (x - a) / math.Pow(TAU, 2)
		return t1 - t2
	}

	A := a
	var B float64

	phiSqPlusV := math.Pow(gCur.Phi, 2) + v
	if math.Pow(delta, 2) > phiSqPlusV {
		B = math.Log(math.Pow(delta, 2) - math.Pow(gCur.Phi, 2) - v)
	} else {
		k := 1
		for f(a-float64(k)*TAU) < 0 {
			k++
		}
		B = a - float64(k)*TAU
	}

	fA := f(A)
	fB := f(B)

	for math.Abs(B-A) > EPSILON {
		C := A + (A-B)*fA/(fB-fA)
		fC := f(C)
		fCfB := fC * fB
		if fCfB <= 0 {
			A = B
			fA = fB
		} else {
			fA = fA / 2
		}
		B = C
		fB = fC
	}

	return math.Exp(A / 2)
}

func getNewRatingDev(gCur Glicko2, sigmaPrime float64) float64 {
	return math.Sqrt(math.Pow(gCur.Phi, 2) + math.Pow(sigmaPrime, 2))
}

func (gCur Glicko2) updateGlicko2Vars(phiStar float64, v float64, gOpponents []Glicko2, scores []float64, sigmaPrime float64) Glicko2 {
	phi := 1 / math.Sqrt(1/math.Pow(phiStar, 2)+1/v)
	mu := gCur.Mu + math.Pow(phi, 2)*computeDelta(gCur, gOpponents, scores)/v
	return Glicko2{
		Mu:    mu,
		Sigma: sigmaPrime,
		Phi:   phi,
	}
}

func (gCur Glicko2) ProcessMatches(gOpps []Glicko2, scores []float64) (Glicko2, error) {
	if len(gOpps) != len(scores) {
		return Glicko2{}, errors.New("arrays must be the same length")
	}
	delta := computeDelta(gCur, gOpps, scores)
	v := computeV(gCur, gOpps)
	sigmaPrime := sigmaByIllinois(gCur, delta, v)
	phiStar := getNewRatingDev(gCur, sigmaPrime)
	return gCur.updateGlicko2Vars(phiStar, v, gOpps, scores, sigmaPrime), nil
}
