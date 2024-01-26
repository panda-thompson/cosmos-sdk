package objcache

import "unsafe"

type KeylessContainer struct {
	m map[string]any
}

func (k KeylessContainer) Set(prefix []byte, value any) {
	_, exists := k.m[unsafeString(prefix)]
	if exists {
		k.m[unsafeString(prefix)] = value
	}
	k.m[string(prefix)] = value
}

func (k KeylessContainer) Get(prefix []byte) (value any, ok bool) {
	value, ok = k.m[unsafeString(prefix)]
	return
}

type NamespacedKeylessContainer struct {
	m map[string]KeylessContainer
}

func NewNamespacedKeylessContainer() NamespacedKeylessContainer {
	return NamespacedKeylessContainer{
		m: make(map[string]KeylessContainer),
	}
}

func (n NamespacedKeylessContainer) GetKeylessContainer(address []byte) KeylessContainer {
	v, ok := n.m[unsafeString(address)]
	if ok {
		return v
	}
	kc := KeylessContainer{m: make(map[string]any)}
	n.m[string(address)] = kc
	return kc
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
