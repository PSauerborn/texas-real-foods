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
    key text NOT NULL,
    description text
);


ALTER TABLE public.access_keys OWNER TO texas_real_foods;

--
-- Name: asset_data; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.asset_data (
    business_id uuid NOT NULL,
    source text NOT NULL,
    website_live boolean DEFAULT true,
    phone text[] DEFAULT '{}'::text[] NOT NULL,
    meta json DEFAULT '{}'::json NOT NULL,
    open boolean DEFAULT true NOT NULL
);


ALTER TABLE public.asset_data OWNER TO texas_real_foods;

--
-- Name: asset_data_timeseries; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.asset_data_timeseries (
    business_id uuid NOT NULL,
    event_timestamp timestamp without time zone DEFAULT now() NOT NULL,
    source text NOT NULL,
    website_live boolean,
    phone text[] DEFAULT '{}'::text[] NOT NULL,
    meta json DEFAULT '{}'::json NOT NULL,
    open boolean DEFAULT true NOT NULL
);


ALTER TABLE public.asset_data_timeseries OWNER TO texas_real_foods;

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
-- Name: mail_relay_entries; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.mail_relay_entries (
    entry_id uuid NOT NULL,
    event_timestamp timestamp without time zone DEFAULT now() NOT NULL,
    status text NOT NULL,
    completed boolean DEFAULT false NOT NULL,
    data json DEFAULT '{}'::json NOT NULL
);


ALTER TABLE public.mail_relay_entries OWNER TO texas_real_foods;

--
-- Name: notification_metadata; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.notification_metadata (
    notification_id uuid NOT NULL,
    metadata json DEFAULT '{}'::json
);


ALTER TABLE public.notification_metadata OWNER TO texas_real_foods;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.notifications (
    notification_id uuid NOT NULL,
    event_timestamp timestamp without time zone DEFAULT now() NOT NULL,
    notification json NOT NULL,
    read boolean DEFAULT false,
    hash text
);


ALTER TABLE public.notifications OWNER TO texas_real_foods;

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
-- Name: asset_data_timeseries asset_data_timeseries_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.asset_data_timeseries
    ADD CONSTRAINT asset_data_timeseries_pkey PRIMARY KEY (business_id, source, event_timestamp);


--
-- Name: asset_metadata asset_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.asset_metadata
    ADD CONSTRAINT asset_metadata_pkey PRIMARY KEY (business_id);


--
-- Name: mail_relay_entries mail_relay_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.mail_relay_entries
    ADD CONSTRAINT mail_relay_entries_pkey PRIMARY KEY (entry_id);


--
-- Name: notification_metadata notification_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.notification_metadata
    ADD CONSTRAINT notification_metadata_pkey PRIMARY KEY (notification_id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (notification_id);


--
-- PostgreSQL database dump complete
--
