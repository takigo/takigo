//go:build darwin

package font

/*
#cgo LDFLAGS: -framework CoreText -framework CoreFoundation
#include <CoreText/CoreText.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

// Helper: get font family name count.
static int ct_family_count(void) {
    CTFontCollectionRef collection = CTFontCollectionCreateFromAvailableFonts(NULL);
    if (!collection) return 0;
    CFArrayRef descriptors = CTFontCollectionCreateMatchingFontDescriptors(collection);
    CFRelease(collection);
    if (!descriptors) return 0;
    int count = (int)CFArrayGetCount(descriptors);
    CFRelease(descriptors);
    return count;
}

// Helper: get all font family names as a NULL-terminated array of C strings.
// Caller must free each string and the array.
static char **ct_list_families(int *outCount) {
    CTFontCollectionRef collection = CTFontCollectionCreateFromAvailableFonts(NULL);
    if (!collection) { *outCount = 0; return NULL; }
    CFArrayRef descriptors = CTFontCollectionCreateMatchingFontDescriptors(collection);
    CFRelease(collection);
    if (!descriptors) { *outCount = 0; return NULL; }

    int count = (int)CFArrayGetCount(descriptors);
    char **result = (char **)calloc(count, sizeof(char *));
    int n = 0;

    for (int i = 0; i < count; i++) {
        CTFontDescriptorRef desc = (CTFontDescriptorRef)CFArrayGetValueAtIndex(descriptors, i);
        CFStringRef family = CTFontDescriptorCopyAttribute(desc, kCTFontFamilyNameAttribute);
        if (!family) continue;

        CFIndex len = CFStringGetLength(family);
        CFIndex maxSize = CFStringGetMaximumSizeForEncoding(len, kCFStringEncodingUTF8) + 1;
        char *buf = (char *)malloc(maxSize);
        if (CFStringGetCString(family, buf, maxSize, kCFStringEncodingUTF8)) {
            result[n++] = buf;
        } else {
            free(buf);
        }
        CFRelease(family);
    }
    CFRelease(descriptors);
    *outCount = n;
    return result;
}
*/
import "C"
import (
	"sort"
	"unsafe"
)

// ListFamilies returns a sorted list of unique available font family names
// using Core Text on macOS.
func ListFamilies() []string {
	var count C.int
	families := C.ct_list_families(&count)
	if families == nil || count == 0 {
		return nil
	}
	defer func() {
		for i := C.int(0); i < count; i++ {
			ptr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(families)) + uintptr(i)*unsafe.Sizeof(families)))
			C.free(unsafe.Pointer(ptr))
		}
		C.free(unsafe.Pointer(families))
	}()

	seen := make(map[string]bool)
	var result []string
	for i := C.int(0); i < count; i++ {
		ptr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(families)) + uintptr(i)*unsafe.Sizeof(families)))
		name := C.GoString(ptr)
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}

	sort.Strings(result)
	return result
}
