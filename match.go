package mux2

func match(pat, str string) (bool, map[string]string) {
	var p, s int
	pr := map[string]string{}
	for {
		if p == len(pat) && s == len(str) {
			// precise match
			return true, pr
		} else if p == len(pat) && p > 0 && pat[p-1] == '/' {
			// pattern ending with /, remaining string
			return true, pr
		} else if p == len(pat) || s == len(str) {
			// running out of pattern or string
			return false, nil
		} else if pat[p] == ':' {
			p0 := p
			s0 := s
			for p != len(pat) && pat[p] != '/' {
				p++
			}
			for s != len(str) && str[s] != '/' {
				s++
			}
			pr[pat[p0+1:p]] = str[s0:s]
		} else if pat[p] != str[s] {
			return false, nil
		} else {
			s++
			p++
		}
	}
}
