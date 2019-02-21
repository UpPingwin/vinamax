package vinamax

import (
	"fmt"
	"log"
	"math"
	"math/rand"
)

//A particle essentially constains a position, magnetisation
type particle struct {
	x, y, z             float64
	m                   vector
	demagnetising_field vector
	u_anis              vector  // Uniaxial anisotropy axis
	u2_anis             vector  // Uniaxial anisotropy axis
	c1_anis             vector  // cubic anisotropy axis
	c2_anis             vector  // cubic anisotropy axis
	c3_anis             vector  // cubic anisotropy axis
	r                   float64 // radius core
	r_h                 float64 // radius core and coating together
	msat                float64 // Saturation magnetisation in A/m
	flip                float64 // time of next flip event
	temp_prefactor          float64
	randomvprefact      float64
	eta                 float64 //viscosity of particle surroundings

	heff           vector //effective field
	dmdt           vector //dm/dt for use in du/dt when condition 1
	omega          vector //for dm/dt and du/dt in condition 1
	tempfield      vector
	randomvfield   vector
	tempm          vector
	previousm      vector
	tempu_anis     vector
	previousu_anis vector
	fehlk1         vector
	fehlk1_u       vector
	fehlk2         vector
	fehlk2_u       vector
	fehlk3         vector
	fehlk3_u       vector
	fehlk4         vector
	fehlk4_u       vector
	fehlk5         vector
	fehlk5_u       vector
	fehlk6         vector
	fehlk6_u       vector
	fehlk7         vector
	fehlk7_u       vector
	fehlk8         vector
	fehlk9         vector
	fehlk10        vector
	fehlk11        vector
	fehlk12        vector
	fehlk13        vector
	fixed	       bool
}

//print position and magnitisation of a particle
func (p particle) string() string {
	return fmt.Sprintf("particle@(%v, %v, %v), %v %v %v", p.x, p.y, p.z, p.m[0], p.m[1], p.m[2])
}

//Gives all particles the same specified uniaxialanisotropy-axis
func Anisotropy_axis(x, y, z float64) {
	uaniscalled = true
	global_u_anis = norm(vector{x, y, z})
	for i := range universe.lijst {
		universe.lijst[i].u_anis = global_u_anis
	}
}

//Gives all particles the same specified cubic1anisotropy-axis
func C1anisotropy_axis(x, y, z float64) {
	c1called = true
	a := norm(vector{x, y, z})
	for i := range universe.lijst {
		universe.lijst[i].c1_anis = a
	}
}

//Gives all particles the same specified second uniaxial anisotropy-axis
func U2anisotropy_axis(x, y, z float64) {
	a := norm(vector{x, y, z})
	for i := range universe.lijst {
		universe.lijst[i].u2_anis = a
	}
}

//Gives all particles the same specified cubic2anisotropy-axis, must be orthogonal to c1
func C2anisotropy_axis(x, y, z float64) {
	c2called = true
	a := norm(vector{x, y, z})
	for i := range universe.lijst {
		if universe.lijst[i].c1_anis.dot(a) == 0 {
			universe.lijst[i].c2_anis = a
			universe.lijst[i].c3_anis = norm(universe.lijst[i].c1_anis.cross(a))
		} else {
			log.Fatal("c1 and c2 should be orthogonal")
		}
	}
}

//Gives all particles a random anisotropy-axis
func Anisotropy_random() {
	uaniscalled = true
	for i := range universe.lijst {
		phi := rng.Float64() * (2 * math.Pi)
		theta := 2 * math.Asin(math.Sqrt(rng.Float64()))
		universe.lijst[i].u_anis = norm(vector{math.Sin(theta) * math.Cos(phi), math.Sin(theta) * math.Sin(phi), math.Cos(theta)})
		if math.Cos(theta) < 0. {
			universe.lijst[i].u_anis = universe.lijst[i].u_anis.times(-1.)
		}
	}
}

//Gives all particles a random anisotropy-axis in the xy plane
func Anisotropy_random_xy() {
	uaniscalled = true
	for i := range universe.lijst {
		phi := rng.Float64() * (2 * math.Pi)
		universe.lijst[i].u_anis = vector{math.Cos(phi), math.Sin(phi), 0}
	}
}

//Gives all particles with random magnetisation orientation
func M_random() {
	magnetisationcalled = true
	for i := range universe.lijst {
		if universe.lijst[i].fixed==false{
		phi := rng.Float64() * (2 * math.Pi)
		theta := 2 * math.Asin(math.Sqrt(rng.Float64()))
		universe.lijst[i].m = vector{math.Sin(theta) * math.Cos(phi), math.Sin(theta) * math.Sin(phi), math.Cos(theta)}
		}
	}
}

//Gives all particles with random magnetisation orientation in the xy plane
func M_random_xy() {
	magnetisationcalled = true
	for i := range universe.lijst {
		if universe.lijst[i].fixed==false{
		phi := rng.Float64() * (2 * math.Pi)
		universe.lijst[i].m = vector{math.Cos(phi), math.Sin(phi), 0}
		}
	}
}

//Gives all particles magnetisation specified by the moment superposition model
func M_MSM(tmag, field float64) {
	r := rand.New(rand.NewSource(99))
	magnetisationcalled = true
	for i := range universe.lijst {
		if universe.lijst[i].fixed==false{
		volume := cube(universe.lijst[i].r) * 4. / 3. * math.Pi
		gprime := Alpha * gamma0 * mu0 / (1. + (Alpha * Alpha))
		delta := Ku1 * volume / (kb * Temp)
		msat := universe.lijst[i].msat
		hk := 2. * Ku1 / (msat * mu0)
		tau0 := gprime * hk * math.Sqrt(delta/math.Pi)
		tauN := 1. / tau0 * math.Exp(Ku1*volume/(kb*Temp)*(1.-0.82*msat*field*mu0/Ku1))
		x := volume * field * msat * mu0 / (kb * Temp)

		langevin := 1./math.Tanh(x) - 1./x

		M := langevin * (1. - math.Exp(-tmag/tauN))
		up := (2.*M + 1.) / (2.) //2.M because of random anisotropy axes
		if r.Float64() < up {
			universe.lijst[i].m = universe.lijst[i].u_anis
		} else {
			universe.lijst[i].m = universe.lijst[i].u_anis.times(-1.)
		}
		}
	}
}

//Gives all particles a specified magnetisation direction
func M_uniform(x, y, z float64) {
	magnetisationcalled = true
	a := norm(vector{x, y, z})
	for i := range universe.lijst {
		if universe.lijst[i].fixed==false{
		universe.lijst[i].m = a
		}
	}
}

//Sets the saturation magnetisation of all particles in A/m
func Msat(x float64) {
	msatcalled = true
	for i := range universe.lijst {
		universe.lijst[i].msat = x
	}
}




//Adds a single particle at specified coordinates with fixed spin, returns false if unsuccesfull
func addfixedparticle(x, y, z, mx, my, mz float64) bool {
	if radiuscalled == false {
		log.Fatal("You have to specify the size of the particles before creating new particles")
	}

	radius := getradius()

	var radius_h float64
	if radius_hcalled == false { //when no hydrodynamic radius is specified, consider it equal to core radius
		radius_h = radius
	}
	if logradiuscalled { //when distribution of core sizes use a fixed coating size
		radius_h = getradius_h() + radius
	}
	if constradiuscalled {
		radius_h = getradius_h()
	}
	if overlap(x, y, z, radius_h) == true {
		return false
	}

	if BrownianRotation == true && viscositycalled == false {
		log.Fatal("You have to specify the viscosity of the particles' surroundings before adding new particles")
	}

	if universe.inworld(vector{x, y, z}) {
		a := particle{x: x, y: y, z: z, r: radius, r_h: radius_h,m:vector{mx,my,mz},fixed:true}
		if BrownianRotation {
			a.eta = viscosity
		}
		universe.lijst = append(universe.lijst, &a)
		universe.number += 1
		msatcalled = false
	} else {
		log.Fatal("Trying to add particle at location (", x, ",", y, ",", z, ") which lies outside of universe")
	}

	return true
}

func Addfixedparticle(x, y, z,mx, my, mz float64) {
	if addfixedparticle(x, y, z,mx,my,mz) == false {
		log.Fatal("Trying to add particle at overlapping locations")
	}
}


