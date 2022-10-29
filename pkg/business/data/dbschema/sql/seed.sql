INSERT INTO public.tests (id, probe_type, frequency, last_run, probe_data) VALUES
	('12345', 'icmp', '30', '1653142252', 'x')
	ON CONFLICT DO NOTHING;
