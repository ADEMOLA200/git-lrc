// Comment component
import { waitForPreact, getBadgeClass, copyToClipboard } from './utils.js';
import { getFeedbackPopup } from './FeedbackPopup.js';

export async function createComment() {
    const { html, useEffect, useState } = await waitForPreact();
    const FeedbackPopup = await getFeedbackPopup();

    const renderFact = (label, value, extraClass = '') => {
        if (!value) {
            return null;
        }
        return html`
            <div class="comment-fact ${extraClass}">
                <span class="comment-fact-label">${label}</span>
                <span class="comment-fact-value">${value}</span>
            </div>
        `;
    };

    return function Comment({ comment, filePath, codeExcerpt, commentId, visibilityKey, isHidden, onToggleVisibility, onFirstRender, renderTimingLabel, vote, onVote }) {
        const [copied, setCopied] = useState(false);

        useEffect(() => {
            if (visibilityKey && onFirstRender) {
                onFirstRender(visibilityKey);
            }
        }, [visibilityKey, onFirstRender]);

        const handleCopy = async (e) => {
            e.stopPropagation();
            
            let copyText = '';
            if (filePath) {
                copyText += filePath;
                if (comment.Line) {
                    copyText += ':' + comment.Line;
                }
                copyText += '\n\n';
            }
            
            if (codeExcerpt) {
                copyText += 'Code excerpt:\n' + codeExcerpt + '\n\n';
            }
            
            copyText += 'Issue:\n' + comment.Content;
            
            try {
                await copyToClipboard(copyText);
                setCopied(true);
                setTimeout(() => setCopied(false), 2000);
            } catch (err) {
                console.error('Copy failed:', err);
            }
        };

        const handleToggleVisibility = (e) => {
            e.stopPropagation();
            if (!visibilityKey) {
                console.warn('Missing visibility key for comment toggle');
                return;
            }
            if (onToggleVisibility) {
                onToggleVisibility(visibilityKey);
            }
        };
        
        const badgeClass = getBadgeClass(comment.Severity);
        const lineLabel = comment.Line ? `:${comment.Line}` : '';
        
        return html`
            <tr class="comment-row ${isHidden ? 'comment-row-hidden' : ''}" data-line="${comment.Line}" id="${commentId}">
                <td colspan="3">
                    <div class="comment-visibility-row">
                        ${isHidden
                            ? html`
                                <div class="comment-hidden-placeholder" style="position: relative;">
                                    <div class="comment-actions" style="display: flex; gap: 8px; position: absolute; right: 12px; top: 12px;">
                                        <button 
                                            class="comment-visibility-btn"
                                            title="Show this comment to the AI Agent"
                                            onClick=${handleToggleVisibility}
                                            style="position: static; opacity: 1;"
                                        >
                                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-7 11-7 11 7 11 7-4 7-11 7S1 12 1 12z"/><circle cx="12" cy="12" r="3"/></svg>
                                            Show
                                        </button>
                                    </div>
                                    <span class="comment-hidden-title">Comment hidden</span>
                                    <span class="comment-hidden-meta">${filePath}${lineLabel}</span>
                                    <span class="comment-hidden-note">Hidden comments are excluded from Copy Visible Issues and the Claude Agent.</span>
                                </div>
                            `
                            : html`
                                <div 
                                    class="comment-container"
                                    data-filepath="${filePath}"
                                    data-line="${comment.Line}"
                                    data-comment="${comment.Content}"
                                >
                                    <div class="comment-actions">
                                        <${FeedbackPopup}
                                            type="up"
                                            vote=${vote}
                                            onVote=${onVote}
                                            visibilityKey=${visibilityKey}
                                            commentContent=${comment.Content}
                                            codeExcerpt=${codeExcerpt}
                                            filePath=${filePath}
                                            severity=${comment.Severity}
                                            sourceType="comment"
                                        />
                                        <${FeedbackPopup}
                                            type="down"
                                            vote=${vote}
                                            onVote=${onVote}
                                            visibilityKey=${visibilityKey}
                                            commentContent=${comment.Content}
                                            codeExcerpt=${codeExcerpt}
                                            filePath=${filePath}
                                            severity=${comment.Severity}
                                            sourceType="comment"
                                        />
                                        <button
                                            class="comment-visibility-btn comment-action-icon-btn"
                                            title="Hide this comment from the AI Agent"
                                            onClick=${handleToggleVisibility}
                                        >
                                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-5 0-9.27-3.11-11-7.5a11.8 11.8 0 012.89-4.11M9.88 9.88a3 3 0 104.24 4.24"/><path d="M1 1l22 22"/></svg>
                                        </button>
                                        <button 
                                            class="comment-copy-btn comment-action-icon-btn ${copied ? 'copied' : ''}"
                                            title="Copy issue details"
                                            onClick=${handleCopy}
                                        >
                                            ${copied
                                                ? html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round"/></svg>`
                                                : html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>`}
                                        </button>
                                    </div>
                                    <div class="comment-header">
                                        <div class="comment-header-main">
                                            <span class="comment-badge ${badgeClass}">${comment.Severity}</span>
                                            <span class="comment-location">${filePath}${lineLabel}</span>
                                            ${renderTimingLabel && html`
                                                <span class="comment-arrival">${renderTimingLabel}</span>
                                            `}
                                        </div>
                                    </div>
                                    <div class="comment-facts-row">
                                        ${renderFact('Confidence', comment.Confidence)}
                                        ${renderFact('Type', comment.Type)}
                                        ${(comment.Category || comment.Subcategory) && html`
                                            <div class="comment-fact comment-fact-classification">
                                                <span class="comment-fact-label">Classification</span>
                                                <span class="comment-fact-value">
                                                    <span>${comment.Category || 'Uncategorized'}</span>
                                                    ${comment.Subcategory && html`
                                                        <span class="comment-fact-separator">/</span>
                                                        <span class="comment-fact-subvalue">${comment.Subcategory}</span>
                                                    `}
                                                </span>
                                            </div>
                                        `}
                                    </div>
                                    <div class="comment-body">${comment.Content}</div>
                                </div>
                            `
                        }
                    </div>
                </td>
            </tr>
        `;
    };
}

let CommentComponent = null;
export async function getComment() {
    if (!CommentComponent) {
        CommentComponent = await createComment();
    }
    return CommentComponent;
}
