import { BellRing } from '@lucide/vue';
import type { ForgeConsolePlugin } from '@fromforgesoftware/forge-console-plugin';
import {
	ResourceListView,
	ResourceCreateForm,
	ActionForm,
} from '@fromforgesoftware/forge-console-plugin/ui';

// PluginContext is what the forge-console-plugin loader passes to a remote
// module's default-export factory. apiBase is resolved at RUNTIME from the
// backend /apps descriptor (descriptor.apiBase), not at build time — this
// module.js is built once, before any deployment knows its gateway base.
export interface PluginContext {
	apiBase: string;
}

// gjallarhornPlugin builds the ForgeConsolePlugin for a given apiBase: a
// delivery-status board, a test-send form (covering the webhook channel +
// scheduledAt), and per-recipient channel opt-out preferences. In the forge
// host this used to call apiBaseFor('gjallarhorn') at construction; in the
// remote module the apiBase is injected by the loader via the factory below.
export function gjallarhornPlugin(apiBase: string): ForgeConsolePlugin {
	return {
		serviceId: 'gjallarhorn',
		type: 'app',
		title: 'Gjallarhorn',
		basePath: '/gjallarhorn',
		apiBase,
		icon: BellRing,
		order: 3,
		pages: [
			{
				path: 'notifications',
				name: 'Delivery board',
				component: ResourceListView,
				props: {
					apiBase,
					type: 'notifications',
					title: 'Delivery board',
					columns: ['recipient', 'channel', 'status', 'subject'],
				},
			},
			{
				path: 'notifications/new',
				name: 'Test send',
				component: ResourceCreateForm,
				props: {
					apiBase,
					type: 'notifications',
					title: 'Test send',
					fields: [
						{ name: 'recipient', label: 'Recipient', required: true },
						{
							name: 'channel',
							label: 'Channel',
							type: 'select',
							options: [
								{ value: 'EMAIL', label: 'Email' },
								{ value: 'WEBHOOK', label: 'Webhook' },
							],
							required: true,
						},
						{ name: 'subject', label: 'Subject' },
						{ name: 'body', label: 'Body' },
						{ name: 'template', label: 'Template' },
						{ name: 'realmId', label: 'Realm ID' },
						{ name: 'scheduledAt', label: 'Scheduled at (RFC3339, blank = now)' },
					],
				},
			},
			{
				path: 'preferences',
				name: 'Set preference',
				component: ActionForm,
				props: {
					apiBase,
					path: '/api/notification-preferences',
					type: 'notification-preferences',
					title: 'Set notification preference',
					submitLabel: 'Save',
					fields: [
						{ name: 'recipient', label: 'Recipient', required: true },
						{
							name: 'channel',
							label: 'Channel',
							type: 'select',
							options: [
								{ value: 'EMAIL', label: 'Email' },
								{ value: 'WEBHOOK', label: 'Webhook' },
							],
							required: true,
						},
						{ name: 'realmId', label: 'Realm ID' },
						{ name: 'suppressed', label: 'Suppressed (mute this channel)', type: 'checkbox' },
					],
				},
			},
		],
	};
}

// Default export: the apiBase-injection FACTORY. A remote module.js is built
// once, before apiBase is known — apiBase only exists at runtime, from the
// backend /apps descriptor. The forge-console-plugin loader calls this factory
// with `{ apiBase: descriptor.apiBase }` and registers the returned plugin.
//
// The factory is also tolerant of being called with no context (the loader's
// zero-arg path) — it falls back to the gateway proxy base the host uses by
// convention so the descriptor fallback in loadConsolePlugins still applies.
export default function createPlugin(ctx?: PluginContext): ForgeConsolePlugin {
	return gjallarhornPlugin(ctx?.apiBase ?? '/api/proxy/gjallarhorn');
}
