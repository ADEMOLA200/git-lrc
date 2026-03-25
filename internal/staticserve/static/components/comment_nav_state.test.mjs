import test from 'node:test';
import assert from 'node:assert/strict';

import {
    normalizeIndex,
    sanitizeCommentNavState,
    reconcileCommentNavState,
    resolveNextIndex,
    resolvePrevIndex
} from './comment_nav_state.mjs';

function buildComments(length) {
    const ids = [];
    for (let i = 0; i < length; i++) ids.push(`C${i}`);
    return ids.map((commentId) => ({ commentId }));
}

function clamp(value, min, max) {
    return Math.min(Math.max(value, min), max);
}

function oracleNormalize(idx, length) {
    if (length <= 0) return -1;
    let out = idx;
    while (out < 0) out += length;
    while (out >= length) out -= length;
    return out;
}

function oracleReconcile(allComments, prevIdx, activeCommentId, anchorIdx) {
    const length = allComments.length;
    if (length === 0) {
        return { currentIdx: -1, activeCommentId: null, anchorIdx: 0 };
    }

    if (activeCommentId !== null) {
        const idx = allComments.findIndex((entry) => entry.commentId === activeCommentId);
        if (idx >= 0) {
            return { currentIdx: idx, activeCommentId, anchorIdx: idx };
        }
    }

    if (prevIdx >= 0) {
        return {
            currentIdx: -1,
            activeCommentId: null,
            anchorIdx: prevIdx <= length ? prevIdx : length
        };
    }

    return {
        currentIdx: -1,
        activeCommentId: null,
        anchorIdx: clamp(anchorIdx, 0, length)
    };
}

function oracleNext(currentIdx, anchorIdx, length) {
    if (length <= 0) return -1;
    return currentIdx >= 0
        ? oracleNormalize(currentIdx + 1, length)
        : oracleNormalize(anchorIdx, length);
}

function oraclePrev(currentIdx, anchorIdx, length) {
    if (length <= 0) return -1;
    return currentIdx >= 0
        ? oracleNormalize(currentIdx - 1, length)
        : oracleNormalize(anchorIdx - 1, length);
}

test('independent structural validation: implementation matches oracle across deletion and focus states', () => {
    for (let length = 0; length <= 6; length++) {
        const allComments = buildComments(length);
        const ids = allComments.map((c) => c.commentId);

        const possiblePrev = [-1, ...ids.map((_, idx) => idx)];
        const possibleAnchor = [];
        for (let a = -2; a <= length + 2; a++) possibleAnchor.push(a);
        const possibleActive = [null, ...ids, 'MISSING'];

        for (const prevIdx of possiblePrev) {
            for (const anchorIdx of possibleAnchor) {
                for (const activeCommentId of possibleActive) {
                    const expected = oracleReconcile(allComments, prevIdx, activeCommentId, anchorIdx);
                    const actual = reconcileCommentNavState(allComments, prevIdx, activeCommentId, anchorIdx);
                    assert.deepEqual(actual, expected);

                    const expectedNext = oracleNext(actual.currentIdx, actual.anchorIdx, length);
                    const expectedPrev = oraclePrev(actual.currentIdx, actual.anchorIdx, length);
                    assert.equal(resolveNextIndex(actual.currentIdx, actual.anchorIdx, length), expectedNext);
                    assert.equal(resolvePrevIndex(actual.currentIdx, actual.anchorIdx, length), expectedPrev);

                    if (length > 0) {
                        assert.equal(normalizeIndex(actual.anchorIdx, length), oracleNormalize(actual.anchorIdx, length));
                    } else {
                        assert.equal(normalizeIndex(actual.anchorIdx, length), -1);
                    }
                }
            }
        }
    }
});

test('hide-current regression: [A,B,C] at B -> hide B -> next lands on C', () => {
    const visibleAfterHide = [{ commentId: 'A' }, { commentId: 'C' }];
    const reconciled = reconcileCommentNavState(visibleAfterHide, 1, 'B', 99);

    assert.equal(reconciled.currentIdx, -1);
    assert.equal(reconciled.anchorIdx, 1);

    const nextIdx = resolveNextIndex(reconciled.currentIdx, reconciled.anchorIdx, visibleAfterHide.length);
    assert.equal(nextIdx, 1);
    assert.equal(visibleAfterHide[nextIdx].commentId, 'C');
});

test('sanitizeCommentNavState bounds malformed state safely', () => {
    const sanitized = sanitizeCommentNavState(
        { currentIdx: 999, activeCommentId: 42, anchorIdx: -5 },
        3
    );

    assert.deepEqual(sanitized, {
        currentIdx: -1,
        activeCommentId: null,
        anchorIdx: 0
    });
});
