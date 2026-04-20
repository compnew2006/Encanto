import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ parent, url }) => {
	const parentData = await parent();
	return {
		user: parentData.user,
		pathname: url.pathname
	};
};

