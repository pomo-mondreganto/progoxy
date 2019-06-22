package main

type Vertex struct {
	next     map[byte]*Vertex
	pred     *Vertex
	chr      byte
	to       map[byte]*Vertex
	link     *Vertex
	terminal bool
	root     bool
}

func (v *Vertex) Add(s []byte) {
	cur := v
	for _, c := range s {
		_, ok := cur.next[c]
		if !ok {
			cur.next[c] = &Vertex{
				pred: cur,
				chr:  c,
				next: make(map[byte]*Vertex, 1),
				to:   make(map[byte]*Vertex, 1),
			}
		}
		cur = cur.next[c]
	}
	cur.terminal = true
}

func (v *Vertex) GetLink() *Vertex {
	if v.link == nil {
		if v.root {
			v.link = v
		} else if v.pred.root {
			v.link = v.pred
		} else {
			v.link = v.pred.GetLink().Go(v.chr)
		}
	}
	return v.link
}

func (v *Vertex) Go(c byte) *Vertex {
	_, ok := v.to[c]
	if !ok {
		nxt, ok := v.next[c]
		if !ok {
			if v.root {
				v.to[c] = v
			} else {
				v.to[c] = v.GetLink().Go(c)
			}
		} else {
			v.to[c] = nxt
		}
	}
	return v.to[c]
}

func (v *Vertex) Find(s []byte) bool {
	cur := v
	for _, c := range s {
		cur = cur.Go(c)
		if cur.terminal {
			return true
		}
	}
	return false
}
