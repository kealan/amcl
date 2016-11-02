/*
Copyright 2015 CertiVox UK Ltd

This file is part of The CertiVox MIRACL IOT Crypto SDK (MiotCL)

MiotCL is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

MiotCL is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with MiotCL.  If not, see <http://www.gnu.org/licenses/>.

You can be released from the requirements of the license by purchasing 
a commercial license.
*/

/* Finite Field arithmetic */
/* CLINT mod p functions */

package main

//import "fmt"

var p BIG=BIG{w: [NLEN]int64(Modulus)}

type FP struct {
	x *BIG
}

/* Constructors */
func NewFPint(a int) *FP {
	F:=new(FP)
	F.x=NewBIGint(a)
	F.nres()
	return F
}

func NewFPbig(a *BIG) *FP {
	F:=new(FP)
	F.x=NewBIGcopy(a)
	F.nres()
	return F
}

func NewFPcopy(a *FP) *FP {
	F:=new(FP)
	F.x=NewBIGcopy(a.x)
	return F
}

func (F *FP) toString() string {
	return F.redc().toString()
}

/* convert to Montgomery n-residue form */
func (F *FP) nres() {
	if MODTYPE!=PSEUDO_MERSENNE && MODTYPE!=GENERALISED_MERSENNE {
		d:=NewDBIGscopy(F.x)
		d.shl(uint(NLEN)*BASEBITS)
		F.x.copy(d.mod(&p))
	}
}

/* convert back to regular form */
func (F *FP) redc() *BIG {
	if MODTYPE!=PSEUDO_MERSENNE && MODTYPE!=GENERALISED_MERSENNE {
		d:=NewDBIGscopy(F.x)
		return mod(d)
	} else {
		r:=NewBIGcopy(F.x)
		return r
	}
}

/* reduce this mod Modulus */
func (F *FP) reduce() {
	F.x.mod(&p)
}

/* test this=0? */
func (F *FP) iszilch() bool {
	F.reduce()
	return F.x.iszilch()
}

/* copy from FP b */
func (F *FP) copy(b *FP ) {
	F.x.copy(b.x)
}

/* set this=0 */
func (F *FP) zero() {
	F.x.zero()
}
	
/* set this=1 */
func (F *FP) one() {
	F.x.one(); F.nres()
}

/* normalise this */
func (F *FP) norm() {
	F.x.norm();
}

/* swap FPs depending on d */
func (F *FP) cswap(b *FP,d int32) {
	F.x.cswap(b.x,d);
}

/* copy FPs depending on d */
func (F *FP) cmove(b *FP,d int32) {
	F.x.cmove(b.x,d)
}

/* this*=b mod Modulus */
func (F *FP) mul(b *FP) {

	F.norm();
	b.norm();
	ea:=EXCESS(F.x)
	eb:=EXCESS(b.x)

	if (ea+1)>FEXCESS/(eb+1) {
		F.reduce()
	}
	d:=mul(F.x,b.x)
	F.x.copy(mod(d))
}

/* this = -this mod Modulus */
func (F *FP) neg() {
	m:=NewBIGcopy(&p)

	F.norm()

	ov:=EXCESS(F.x); 
	sb:=uint(1); for ov!=0 {sb++;ov>>=1} 

	m.fshl(sb)
	F.x.rsub(m)		

	if EXCESS(F.x)>=FEXCESS {F.reduce()}
}


/* this*=c mod Modulus, where c is a small int */
func (F *FP) imul(c int) {
	F.norm()
	s:=false
	if (c<0) {
		c=-c
		s=true
	}
	afx:=(EXCESS(F.x)+1)*(int64(c)+1)+1;
	if (c<NEXCESS && afx<FEXCESS) {
		F.x.imul(c);
	} else {
		if (afx<FEXCESS) {
			F.x.pmul(c)
		} else {
			d:=F.x.pxmul(c)
			F.x.copy(d.mod(&p))
		}
	}
	if s {F.neg()}
	F.norm()
}

/* this*=this mod Modulus */
func (F *FP) sqr() {
	F.norm();
	ea:=EXCESS(F.x)
	if (ea+1)>FEXCESS/(ea+1) {
		F.reduce()
	}
	d:=sqr(F.x)	
	F.x.copy(mod(d))
}

/* this+=b */
func (F *FP) add(b *FP) {
	F.x.add(b.x)
	if (EXCESS(F.x)+2>=FEXCESS) {F.reduce()}
}

/* this-=b */
func (F *FP) sub(b *FP) {
	n:=NewFPcopy(b)
	n.neg()
	F.add(n)
}

/* this/=2 mod Modulus */
func (F *FP) div2() {
	F.x.norm()
	if (F.x.parity()==0) {
		F.x.fshr(1)
	} else {
		F.x.add(&p)
		F.x.norm()
		F.x.fshr(1)
	}
}

/* this=1/this mod Modulus */
func (F *FP) inverse() {
	r:=F.redc()
	r.invmodp(&p)
	F.x.copy(r)
	F.nres()
}

/* return TRUE if this==a */
func (F *FP) equals(a *FP) bool {
	a.reduce()
	F.reduce()
	if (comp(a.x,F.x)==0) {return true}
	return false
}

/* return this^e mod Modulus */
func (F *FP) pow(e *BIG) *FP {
	r:=NewFPint(1)
	e.norm()
	F.x.norm()
	m:=NewFPcopy(F)
	for true {
		bt:=e.parity();
		e.fshr(1);
		if bt==1 {r.mul(m)}
		if e.iszilch() {break}
		m.sqr();
	}
	r.x.mod(&p);
	return r;
}

/* return sqrt(this) mod Modulus */
func (F *FP) sqrt() *FP {
	F.reduce();
	b:=NewBIGcopy(&p)
	if MOD8==5 {
		b.dec(5); b.norm(); b.shr(3)
		i:=NewFPcopy(F); i.x.shl(1)
		v:=i.pow(b)
		i.mul(v); i.mul(v)
		i.x.dec(1)
		r:=NewFPcopy(F)
		r.mul(v); r.mul(i) 
		r.reduce()
		return r
	} else {
		b.inc(1); b.norm(); b.shr(2)
		return F.pow(b);
	}
}

/* return jacobi symbol (this/Modulus) */
func (F *FP) jacobi() int {
	w:=F.redc();
	return w.jacobi(&p)
}
