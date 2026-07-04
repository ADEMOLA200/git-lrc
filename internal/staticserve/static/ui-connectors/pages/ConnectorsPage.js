import { renderIcon } from '../../components/icons.js';

const { html, useEffect, useState } = window.preact;

function formatTimestamp(value) {
  if (!value) return '';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return String(value);
  return parsed.toLocaleString();
}

function parseTimestamp(value) {
  if (!value) return null;
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return null;
  return parsed;
}

function formatRelativeTime(value) {
  const dateValue = parseTimestamp(value);
  if (!dateValue) return '—';

  const deltaMs = Date.now() - dateValue.getTime();
  const absMs = Math.abs(deltaMs);
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;
  const week = 7 * day;
  const month = 30 * day;
  const year = 365 * day;

  let amount = 0;
  let unit = 'minute';

  if (absMs >= year) {
    amount = Math.round(deltaMs / year);
    unit = 'year';
  } else if (absMs >= month) {
    amount = Math.round(deltaMs / month);
    unit = 'month';
  } else if (absMs >= week) {
    amount = Math.round(deltaMs / week);
    unit = 'week';
  } else if (absMs >= day) {
    amount = Math.round(deltaMs / day);
    unit = 'day';
  } else if (absMs >= hour) {
    amount = Math.round(deltaMs / hour);
    unit = 'hour';
  } else {
    amount = Math.round(deltaMs / minute);
    unit = 'minute';
  }

  const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });
  return formatter.format(-amount, unit);
}

export function ConnectorsPage({
  connectors,
  loading,
  orderedConnectors,
  hasOrderChanges,
  activeRoleTab,
  onSwitchRoleTab,
  roleCounts,
  helperSettings,
  helperSettingsSaving,
  helperSettingsSaved,
  helperSettingsError,
  onHelperSettingChange,
  onSaveHelperSettings,
  onRefresh,
  onSaveOrder,
  onMove,
  onDragReorder,
  onEdit,
  onDelete,
  onAdd,
}) {
  const [draggingID, setDraggingID] = useState('');
  const [dragOverID, setDragOverID] = useState('');
  const [pulseItemID, setPulseItemID] = useState('');
  const [pulseSave, setPulseSave] = useState(false);
  const [infoExpanded, setInfoExpanded] = useState(false);

  useEffect(() => {
    if (!pulseItemID) {
      return;
    }
    const timer = window.setTimeout(() => setPulseItemID(''), 900);
    return () => window.clearTimeout(timer);
  }, [pulseItemID]);

  useEffect(() => {
    if (!pulseSave) {
      return;
    }
    const timer = window.setTimeout(() => setPulseSave(false), 900);
    return () => window.clearTimeout(timer);
  }, [pulseSave]);

  function handleDragStart(event, connectorID) {
    const id = String(connectorID);
    setDraggingID(id);
    setDragOverID('');
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', id);
    }
  }

  function handleDragOver(event, connectorID) {
    event.preventDefault();
    if (draggingID && draggingID !== String(connectorID)) {
      setDragOverID(String(connectorID));
    }
  }

  function handleDrop(event, connectorID) {
    event.preventDefault();
    const sourceID = event.dataTransfer?.getData('text/plain') || draggingID;
    const targetID = String(connectorID);

    if (sourceID && targetID && sourceID !== targetID) {
      setPulseItemID(sourceID);
      setPulseSave(true);
    }

    onDragReorder(sourceID, targetID);
    setDraggingID('');
    setDragOverID('');
  }

  function handleDragEnd() {
    setDraggingID('');
    setDragOverID('');
  }

  return html`
    <div class="single">
      <section class="card">
        <h2>AI Connectors (${connectors.length})</h2>
        <div class="connectors-content">
          <div class="toolbar connectors-toolbar">
            <div class="btn-group btn-group-compact">
              <button class="secondary" onClick=${onRefresh} disabled=${loading}>
                ${renderIcon(html, 'refresh', { className: `btn-icon ${loading ? 'ui-icon-spin' : ''}` })}${loading ? 'Refreshing...' : 'Refresh'}
              </button>
              <button class=${`${hasOrderChanges ? '' : 'secondary'} ${pulseSave ? 'pulse-attention' : ''}`.trim()} onClick=${onSaveOrder} disabled=${orderedConnectors.length < 2 || !hasOrderChanges}>
                ${renderIcon(html, 'reorder', { className: 'btn-icon' })}Save Priority
              </button>
            </div>
            <div class="btn-group">
              <button onClick=${onAdd}>
                ${renderIcon(html, 'add', { className: 'btn-icon' })}Add Connector
              </button>
            </div>
          </div>

          <div class="role-tabs">
            <button
              class=${`role-tab ${activeRoleTab === 'leader' ? 'active' : ''}`}
              onClick=${() => onSwitchRoleTab('leader')}
            >
              Leader <span class="role-tab-count">${roleCounts.leader}</span>
            </button>
            <button
              class=${`role-tab ${activeRoleTab === 'helper' ? 'active' : ''}`}
              onClick=${() => onSwitchRoleTab('helper')}
            >
              Helper <span class="role-tab-count">${roleCounts.helper}</span>
            </button>
          </div>

          ${activeRoleTab === 'helper'
            ? html`
                <div class="card helper-info-card">
                  <button type="button" class="helper-info-toggle" onClick=${() => setInfoExpanded(!infoExpanded)}>
                    <h2>What is Adaptive Review?</h2>
                    ${renderIcon(html, infoExpanded ? 'dropdownOpen' : 'dropdownClosed', { className: 'btn-icon' })}
                  </button>
                  ${infoExpanded
                    ? html`
                        <div class="helper-info-body">
                          <p class="muted helper-info-point">
                            ${renderIcon(html, 'info', { className: 'btn-icon helper-info-icon' })}
                            <span>Adaptive Review pairs a Leader model (finds and judges issues) with a Helper model (expands the Leader's short notes into clear, polished comments). Splitting the work this way typically cuts review cost 40-50% with no loss in detection quality, since the Leader still decides everything about what's worth flagging.</span>
                          </p>
                          <p class="muted helper-info-point">
                            ${renderIcon(html, 'info', { className: 'btn-icon helper-info-icon' })}
                            <span>If the Helper model fails or isn't configured, LiveReview automatically falls back to posting the Leader model's own output — reviews never fail because of a Helper model issue.</span>
                          </p>
                          <p class="muted helper-info-point" style="margin-bottom: 0;">
                            ${renderIcon(html, 'info', { className: 'btn-icon helper-info-icon' })}
                            <span><strong>Concise Then Expand</strong> asks the Leader for terse notes and has the Helper expand them into full comments. <strong>Polish Only</strong> asks the Leader for full comments and has the Helper just clean up the wording.</span>
                          </p>
                        </div>
                      `
                    : ''}
                </div>

                <div class="card helper-settings-card">
                  <div class="helper-settings-header">
                    <h2>Helper Model Settings</h2>
                    <span class=${`badge ${helperSettings.helperEnabled ? 'badge-enabled' : ''}`}>
                      ${helperSettings.helperEnabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>

                  <label class="checkbox-row">
                    <input
                      type="checkbox"
                      checked=${helperSettings.helperEnabled}
                      onChange=${(event) => onHelperSettingChange('helperEnabled', event.target.checked)}
                    />
                    Enable Helper model for text expansion and polishing
                  </label>

                  <label>Helper Mode</label>
                  <select
                    value=${helperSettings.helperMode}
                    disabled=${!helperSettings.helperEnabled}
                    onChange=${(event) => onHelperSettingChange('helperMode', event.target.value)}
                  >
                    <option value="concise_then_expand">Concise Then Expand</option>
                    <option value="polish_only">Polish Only</option>
                  </select>

                  <div class="row helper-settings-actions">
                    <button onClick=${onSaveHelperSettings} disabled=${helperSettingsSaving}>
                      ${renderIcon(html, 'save', { className: 'btn-icon' })}${helperSettingsSaving ? 'Saving...' : 'Save Helper Settings'}
                    </button>
                    ${helperSettingsSaved ? html`<span class="status ok">Saved</span>` : ''}
                  </div>
                  ${helperSettingsError ? html`<div class="status err">${helperSettingsError}</div>` : ''}
                </div>
              `
            : ''}

          <h3 class="connectors-list-title">Your Connectors</h3>
          <div class="list">
            ${orderedConnectors.length === 0
              ? html`<div class="page-empty">No ${activeRoleTab} connectors found.</div>`
              : orderedConnectors.map((connector, index) => html`
                  <div
                    class=${`item ${draggingID === String(connector.id) ? 'dragging' : ''} ${dragOverID === String(connector.id) ? 'drag-over' : ''} ${pulseItemID === String(connector.id) ? 'pulse-attention' : ''}`}
                    onDragOver=${(event) => handleDragOver(event, connector.id)}
                    onDrop=${(event) => handleDrop(event, connector.id)}
                  >
                    <div class="connector-row">
                      <div class="connector-main">
                        <span class="item-title">${connector.connector_name || 'Connector'}</span>
                        <span class="badge badge-id">#${connector.id}</span>
                        <span class="badge">${connector.provider_name}</span>
                      </div>
                      <div class="row connector-actions">
                        <button
                          class="secondary icon-only drag-handle"
                          title="Drag to reorder"
                          draggable="true"
                          onDragStart=${(event) => handleDragStart(event, connector.id)}
                          onDragEnd=${handleDragEnd}
                        >
                          ${renderIcon(html, 'drag', { className: 'btn-icon' })}
                        </button>
                        <button class="secondary icon-only" title="Move up" disabled=${index === 0} onClick=${() => onMove(String(connector.id), 'up')}>
                          ${renderIcon(html, 'moveUp', { className: 'btn-icon' })}
                        </button>
                        <button class="secondary icon-only" title="Move down" disabled=${index === orderedConnectors.length - 1} onClick=${() => onMove(String(connector.id), 'down')}>
                          ${renderIcon(html, 'moveDown', { className: 'btn-icon' })}
                        </button>
                        <button class="secondary" onClick=${() => onEdit(connector)}>
                          ${renderIcon(html, 'edit', { className: 'btn-icon' })}Edit
                        </button>
                        <button class="tertiary-danger" onClick=${() => onDelete(connector.id)}>
                          ${renderIcon(html, 'delete', { className: 'btn-icon' })}Delete
                        </button>
                      </div>
                    </div>
                    <div class="connector-foot">
                      <div class="muted">Priority #${index + 1}${connector.selected_model ? ` · model: ${connector.selected_model}` : ''}</div>
                      <div
                        class="muted muted-meta muted-meta-right"
                        title=${(() => {
                          const created = formatTimestamp(connector.created_at || connector.createdAt);
                          const updated = formatTimestamp(connector.updated_at || connector.updatedAt);
                          if (created && updated) {
                            return `Added: ${created}\nUpdated: ${updated}`;
                          }
                          if (updated) {
                            return `Updated: ${updated}`;
                          }
                          if (created) {
                            return `Added: ${created}`;
                          }
                          return 'Timestamp unavailable';
                        })()}
                      >
                        ${formatRelativeTime(connector.created_at || connector.createdAt || connector.updated_at || connector.updatedAt)}
                      </div>
                    </div>
                  </div>
                `)}
          </div>
        </div>
      </section>
    </div>
  `;
}
