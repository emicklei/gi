package internal

import "fmt"

type LinuxFilePermissions int

// user,group,other rwx permissions
// rw-r--r-- = 0644
// rwxr-xr-x = 0755
// rwsr-xr-x = 4755
// rwxrwxrwx = 0777
func MustParse(permissions string) LinuxFilePermissions {
	if len(permissions) != 9 {
		panic(fmt.Sprintf("permissions string must be 9 characters, got %d", len(permissions)))
	}

	var p LinuxFilePermissions

	// User
	if permissions[0] == 'r' {
		p |= 0400
	}
	if permissions[1] == 'w' {
		p |= 0200
	}
	if permissions[2] == 'x' {
		p |= 0100
	} else if permissions[2] == 's' {
		p |= 04100 // setuid (04000) + execute (0100)
	} else if permissions[2] == 'S' {
		p |= 04000 // setuid only
	}

	// Group
	if permissions[3] == 'r' {
		p |= 0040
	}
	if permissions[4] == 'w' {
		p |= 0020
	}
	if permissions[5] == 'x' {
		p |= 0010
	} else if permissions[5] == 's' {
		p |= 02010 // setgid (02000) + execute (0010)
	} else if permissions[5] == 'S' {
		p |= 02000 // setgid only
	}

	// Other
	if permissions[6] == 'r' {
		p |= 0004
	}
	if permissions[7] == 'w' {
		p |= 0002
	}
	if permissions[8] == 'x' {
		p |= 0001
	} else if permissions[8] == 't' {
		p |= 01001 // sticky (01000) + execute (0001)
	} else if permissions[8] == 'T' {
		p |= 01000 // sticky only
	}

	return p
}

func (p LinuxFilePermissions) String() string {
	perm := []rune{'-', '-', '-', '-', '-', '-', '-', '-', '-'}
	// User
	if p&0400 != 0 {
		perm[0] = 'r'
	}
	if p&0200 != 0 {
		perm[1] = 'w'
	}
	if p&04000 != 0 {
		if p&0100 != 0 {
			perm[2] = 's'
		} else {
			perm[2] = 'S'
		}
	} else if p&0100 != 0 {
		perm[2] = 'x'
	}

	// Group
	if p&0040 != 0 {
		perm[3] = 'r'
	}
	if p&0020 != 0 {
		perm[4] = 'w'
	}
	if p&02000 != 0 {
		if p&0010 != 0 {
			perm[5] = 's'
		} else {
			perm[5] = 'S'
		}
	} else if p&0010 != 0 {
		perm[5] = 'x'
	}

	// Other
	if p&0004 != 0 {
		perm[6] = 'r'
	}
	if p&0002 != 0 {
		perm[7] = 'w'
	}
	if p&01000 != 0 {
		if p&0001 != 0 {
			perm[8] = 't'
		} else {
			perm[8] = 'T'
		}
	} else if p&0001 != 0 {
		perm[8] = 'x'
	}

	return string(perm)
}
