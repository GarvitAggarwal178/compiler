package integration_test

func arithmeticCases() []compilerCase {
	return []compilerCase{
		{"arithmetic_cycle", `.decl p(a:number,b:number)
.decl q(a:number,b:number)
.decl out(a:number,b:number)
.output out
p(2,2).
q(3,1).
out(x,y) :- p(x+1,y),q(y+1,x).
`, nil},
		{"arithmetic_cycle3", `.decl p(a:number,b:number)
.decl q(a:number,b:number)
.decl r(a:number,b:number)
.decl out(a:number,b:number,c:number)
.output out
p(2,2).
q(3,3).
r(4,1).
out(x,y,z) :- p(x+1,y),q(y+1,z),r(z+1,x).
`, nil},
		{"arithmetic_keyed_query", `.decl p(a:number,b:number)
.decl q(a:number,b:number)
.decl rel(k:number,a:number,b:number)
.decl out(a:number,b:number)
.output out
p(2,2).
q(3,1).
rel(k,x,y) :- p(x+1,y),q(y+1,x),k=7.
out(x,y) :- rel(7,x,y).
`, nil},
		{"signed_query", `.decl p(a:number,b:number)
.decl out(b:number)
.output out
p(-1,2).
out(y) :- p(-1,y).
`, nil},
	}
}
