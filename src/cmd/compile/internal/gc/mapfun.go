package gc

func deindexmap(fn *Node) {
	savefn := Curfn
	Curfn = fn
	deindexmapnode(&fn)
	if fn != Curfn {
		Fatal("deindexmapnode replaced curfn")
	}
	Curfn = savefn
}

func deindexmapnode(np **Node) {
	if *np == nil {
		return
	}

	n := *np

	deindexmapnodelist(n.Ninit)

	deindexmapnode(&n.Left)
	if n.Left != nil && n.Left.Op == OINDEXMAPTMP {
		deindexmapconv(&n.Left)
	}

	deindexmapnode(&n.Right)
	if n.Right != nil && n.Right.Op == OINDEXMAPTMP {
		deindexmapconv(&n.Right)
	}

	deindexmapnodelist(n.List)
	for l := n.List; l != nil; l = l.Next {
		if l.N.Op == OINDEXMAPTMP {
			deindexmapconv(&l.N)
		}
	}

	deindexmapnodelist(n.Rlist)
	for l := n.Rlist; l != nil; l = l.Next {
		if l.N.Op == OINDEXMAPTMP {
			deindexmapconv(&l.N)
		}
	}

	deindexmapnodelist(n.Nbody)

	if n.Op == OINDEXMAP {
		mkindexmaptmp(np)
	}
}

func deindexmapnodelist(l *NodeList) {
	for ; l != nil; l = l.Next {
		deindexmapnode(&l.N)
	}
}

func deindexmapconv(np **Node) {
	n := *np
	r := n.Rlist.N
	addinit(&r, concat(n.Ninit, n.Nbody))
	*np = r
}

// On return:
// ninit has retvar initialization
// nbody does the access
// rlist contain input, output parameters
func mkindexmaptmp(np **Node) {
	n := *np
	indexmaptmp := Nod(OINDEXMAPTMP, nil, nil)

	//	println("hello", Jconv(n, 0))

	t := n.Left.Type
	retval := temp(t.Type)
	indexmaptmp.Ninit = list(indexmaptmp.Ninit, Nod(ODCL, retval, nil))

	var r *Node

	p := ""
	if t.Type.Width <= 128 { // Check ../../runtime/hashmap.go:maxValueSize before changing.
		switch Simsimtype(t.Down) {
		case TINT32, TUINT32:
			p = "mapaccess1_fast32"

		case TINT64, TUINT64:
			p = "mapaccess1_fast64"

		case TSTRING:
			p = "mapaccess1_faststr"
		}
	}

	var key *Node
	if p != "" {
		// fast versions take key by value
		key = n.Right
	} else {
		// standard version takes key by reference.
		// orderexpr made sure key is addressable.
		key = Nod(OADDR, n.Right, nil)

		p = "mapaccess1"
	}

	call := Nod(OCALL, mapfn(p, t), nil)
	call.List = list(call.List, typename(t))
	call.List = list(call.List, n.Left)
	call.List = list(call.List, key)

	if t.Type.Width <= 256 && (Debug['T'] == 0 || t.Type.Width <= 4) {
		return
		// *p(key)
		r = Nod(OIND, call, nil)
		r.Type = t.Type
		r.Typecheck = 1

		indexmaptmp.Nbody = list(indexmaptmp.Nbody, Nod(OAS, retval, r))
	} else {
		// if tmp2 := p(key); tmp2 != runtime.zeroptr {
		//     retval = *tmp2
		// }

		tmp2 := temp(Ptrto(t.Type))
		nif := Nod(OIF, nil, nil)
		nas := Nod(OAS, tmp2, call)
		nas.Colas = true
		nif.Ninit = list1(nas)
		z := Nod(OCONV, syslook("zeroptr", 1), nil)
		z.Type = Ptrto(t.Type)
		nif.Left = Nod(ONE, tmp2, z)
		nif.Nbody = list(list1(Nod(OAS, retval, Nod(OIND, tmp2, nil))), Nod(OVARKILL, tmp2, nil))
		typecheck(&nif, Etop)
		indexmaptmp.Nbody = list(indexmaptmp.Nbody, nif)
	}
	indexmaptmp.Rlist = list(indexmaptmp.Rlist, retval)
	*np = indexmaptmp
}
