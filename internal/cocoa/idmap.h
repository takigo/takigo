// idmap.h: a small open-addressing hash map from nonzero integer IDs to
// pointers, for the Cocoa backend's GC and pixmap registries. They are
// looked up on every drawing primitive, where boxing each ID in an
// NSNumber for an NSDictionary lookup cost an allocation per call.
// Loop-thread only, like the registries it replaces.

#ifndef TAKIGO_IDMAP_H
#define TAKIGO_IDMAP_H

#include <stdint.h>
#include <stdlib.h>

#define IDMAP_TOMB UINT64_MAX

typedef struct {
    uint64_t *keys; // 0 = empty, IDMAP_TOMB = deleted
    void **vals;
    size_t cap;     // power of two, or 0 before the first put
    size_t used;    // live entries plus tombstones
    size_t live;
} idmap;

static inline size_t idmap_slot(uint64_t k, size_t cap) {
    k ^= k >> 33;
    k *= 0xff51afd7ed558ccdULL;
    k ^= k >> 33;
    return (size_t)k & (cap - 1);
}

static inline void *idmap_get(const idmap *m, uint64_t k) {
    if (m->cap == 0 || k == 0 || k == IDMAP_TOMB) return NULL;
    for (size_t i = idmap_slot(k, m->cap);; i = (i + 1) & (m->cap - 1)) {
        uint64_t cur = m->keys[i];
        if (cur == k) return m->vals[i];
        if (cur == 0) return NULL;
    }
}

static void idmap_put(idmap *m, uint64_t k, void *v);

static int idmap_grow(idmap *m) {
    size_t ncap = m->cap ? m->cap * 2 : 16;
    if (m->live * 2 < m->cap) ncap = m->cap; // mostly tombstones: rehash in place
    idmap old = *m;
    m->keys = (uint64_t *)calloc(ncap, sizeof(uint64_t));
    m->vals = (void **)calloc(ncap, sizeof(void *));
    if (!m->keys || !m->vals) {
        free(m->keys);
        free(m->vals);
        *m = old;
        return 0;
    }
    m->cap = ncap;
    m->used = m->live = 0;
    for (size_t i = 0; i < old.cap; i++) {
        if (old.keys[i] != 0 && old.keys[i] != IDMAP_TOMB) idmap_put(m, old.keys[i], old.vals[i]);
    }
    free(old.keys);
    free(old.vals);
    return 1;
}

static void idmap_put(idmap *m, uint64_t k, void *v) {
    if (k == 0 || k == IDMAP_TOMB) return;
    if ((m->used + 1) * 4 > m->cap * 3 && !idmap_grow(m)) return;
    size_t tomb = (size_t)-1;
    for (size_t i = idmap_slot(k, m->cap);; i = (i + 1) & (m->cap - 1)) {
        uint64_t cur = m->keys[i];
        if (cur == k) {
            m->vals[i] = v;
            return;
        }
        if (cur == IDMAP_TOMB && tomb == (size_t)-1) tomb = i;
        if (cur == 0) {
            if (tomb != (size_t)-1) {
                i = tomb; // reuse a tombstone; used already counts it
            } else {
                m->used++;
            }
            m->keys[i] = k;
            m->vals[i] = v;
            m->live++;
            return;
        }
    }
}

// idmap_del removes k and returns its value, or NULL if absent.
static inline void *idmap_del(idmap *m, uint64_t k) {
    if (m->cap == 0 || k == 0 || k == IDMAP_TOMB) return NULL;
    for (size_t i = idmap_slot(k, m->cap);; i = (i + 1) & (m->cap - 1)) {
        uint64_t cur = m->keys[i];
        if (cur == k) {
            void *v = m->vals[i];
            m->keys[i] = IDMAP_TOMB;
            m->vals[i] = NULL;
            m->live--;
            return v;
        }
        if (cur == 0) return NULL;
    }
}

#endif
