package mapper

func New[L, R comparable](ltr map[L]R) M[L, R] {
	rtl := make(map[R]L, len(ltr))
	for k, v := range ltr {
		rtl[v] = k
	}

	return M[L, R]{ltr: ltr, rtl: rtl}
}

type M[L, R comparable] struct {
	ltr map[L]R
	rtl map[R]L
}

func (m M[L, R]) L() map[L]R {
	return m.ltr
}

func (m M[L, R]) R() map[R]L {
	return m.rtl
}
