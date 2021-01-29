--
-- PostgreSQL database dump
--

-- Dumped from database version 12.4 (Debian 12.4-1.pgdg100+1)
-- Dumped by pg_dump version 12.4 (Debian 12.4-1.pgdg100+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: access_keys; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.access_keys (
    key text NOT NULL
);


ALTER TABLE public.access_keys OWNER TO texas_real_foods;

--
-- Name: asset_data; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.asset_data (
    business_id uuid NOT NULL,
    source text NOT NULL,
    website_live boolean DEFAULT true,
    phone text[] DEFAULT '{}'::text[] NOT NULL
);


ALTER TABLE public.asset_data OWNER TO texas_real_foods;

--
-- Name: asset_metadata; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.asset_metadata (
    business_id uuid NOT NULL,
    business_name text NOT NULL,
    last_update timestamp without time zone DEFAULT now() NOT NULL,
    added timestamp without time zone DEFAULT now() NOT NULL,
    metadata json DEFAULT '{}'::json,
    uri text NOT NULL
);


ALTER TABLE public.asset_metadata OWNER TO texas_real_foods;

--
-- Name: asset_updates; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.asset_updates (
    business_id uuid NOT NULL,
    fields_updated text[] NOT NULL,
    "timestamp" timestamp without time zone NOT NULL
);


ALTER TABLE public.asset_updates OWNER TO texas_real_foods;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.notifications (
    notification_id uuid NOT NULL,
    event_timestamp timestamp without time zone DEFAULT now() NOT NULL,
    notification json NOT NULL,
    read boolean DEFAULT false
);


ALTER TABLE public.notifications OWNER TO texas_real_foods;

--
-- Data for Name: access_keys; Type: TABLE DATA; Schema: public; Owner: texas_real_foods
--

COPY public.access_keys (key) FROM stdin;
\.


--
-- Data for Name: asset_data; Type: TABLE DATA; Schema: public; Owner: texas_real_foods
--

COPY public.asset_data (business_id, source, website_live, phone) FROM stdin;
\.


--
-- Data for Name: asset_metadata; Type: TABLE DATA; Schema: public; Owner: texas_real_foods
--

COPY public.asset_metadata (business_id, business_name, last_update, added, metadata, uri) FROM stdin;
9a2cae53-1104-4688-b3d9-53953f23f003	Texas Real Foods	2021-01-29 07:35:49.705531	2021-01-27 20:02:38.978632	null	https://help.texasrealfood.com/support/solutions
da221bcd-7083-426e-a181-4c7e4dee884f	Texas Coffee Traders	2021-01-29 07:41:44.445789	2021-01-27 20:21:05.56192	{"yelp_business_id":"texas-coffee-traders-austin"}	https://www.texascoffeetraders.com/
5c7a40b9-155e-4b9f-acff-3afc4ab12b9e	Boggy Creek Farm	2021-01-29 07:41:44.458724	2021-01-27 20:23:37.720543	{"yelp_business_id":"boggy-creek-farm-austin"}	https://www.boggycreekfarm.com/
42e15fa3-c07f-46c8-88ea-b42b38ad352d	OMG Squee	2021-01-29 07:41:44.468695	2021-01-27 20:26:47.731701	{"yelp_business_id":"omg-squee-austin"}	https://www.squeeclub.com/
1af97836-53e5-468d-9afc-fe6d7d6a521f	Longhorn Meat Market	2021-01-29 07:41:44.480663	2021-01-27 20:28:15.63467	{"yelp_business_id":"longhorn-meat-market-austin"}	https://longhornmeatmarket.com/
3dcb6355-aa5e-47ce-9813-12a828981fc3	Salt & Time	2021-01-29 07:41:44.510585	2021-01-27 20:28:46.317563	{"yelp_business_id":"salt-and-time-austin-6"}	http://www.saltandtime.com/
\.


--
-- Data for Name: asset_updates; Type: TABLE DATA; Schema: public; Owner: texas_real_foods
--

COPY public.asset_updates (business_id, fields_updated, "timestamp") FROM stdin;
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: texas_real_foods
--

COPY public.notifications (notification_id, event_timestamp, notification, read) FROM stdin;
\.


--
-- Name: access_keys access_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.access_keys
    ADD CONSTRAINT access_keys_pkey PRIMARY KEY (key);


--
-- Name: asset_data asset_data_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.asset_data
    ADD CONSTRAINT asset_data_pkey PRIMARY KEY (business_id, source);


--
-- Name: asset_updates asset_updates_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.asset_updates
    ADD CONSTRAINT asset_updates_pkey PRIMARY KEY (business_id);


--
-- Name: asset_metadata assets_meta_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.asset_metadata
    ADD CONSTRAINT assets_meta_pkey PRIMARY KEY (business_id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (notification_id);


--
-- PostgreSQL database dump complete
--

