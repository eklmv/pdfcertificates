CREATE TABLE IF NOT EXISTS template (
    template_id serial PRIMARY KEY,
    content text NOT NULL CHECK (content != '')
);

CREATE TABLE IF NOT EXISTS course (
    course_id serial PRIMARY KEY,
    data jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS student (
    student_id serial PRIMARY KEY,
    data jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS certificate (
    certificate_id char(8) PRIMARY KEY,
    template_id integer NOT NULL REFERENCES template ON DELETE RESTRICT,
    course_id integer NOT NULL REFERENCES course ON DELETE RESTRICT,
    student_id integer NOT NULL REFERENCES student ON DELETE CASCADE,
    timestamp timestamptz NOT NULL DEFAULT now(),
    data jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION gen_id() RETURNS trigger AS $gen_id$
    DECLARE
        new_id char(8);
        counter integer := 0;
        n_attempts CONSTANT integer := 10;
    BEGIN
        LOOP
            new_id = encode(gen_random_bytes(4), 'hex');
            IF (SELECT certificate_id FROM certificate WHERE certificate_id=new_id) IS NULL THEN
                NEW.certificate_id = new_id;
                RETURN NEW;
            END IF;
            counter := counter + 1;
            IF counter > 10 THEN
                RAISE EXCEPTION 'certificate_id generation failed, retry limit exceeded';
                EXIT;
            END IF;
        END LOOP;
    END;
$gen_id$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER gen_id BEFORE INSERT ON certificate
FOR EACH ROW EXECUTE FUNCTION gen_id();

CREATE OR REPLACE FUNCTION update_timestamp() RETURNS trigger AS $update_timestamp$
BEGIN
    IF TG_TABLE_NAME = 'certificate' THEN
        IF NEW.timestamp = OLD.timestamp THEN
            NEW.timestamp := now();
        END IF;
    ELSIF TG_TABLE_NAME = 'student' THEN
        UPDATE certificate SET timestamp = now() WHERE student_id = NEW.student_id;
    ELSIF TG_TABLE_NAME = 'course' THEN
        UPDATE certificate SET timestamp = now() WHERE course_id = NEW.course_id;
    ELSIF TG_TABLE_NAME = 'template' THEN
        UPDATE certificate SET timestamp = now() WHERE template_id = NEW.template_id;
    END IF;
    RETURN NEW;
END;
$update_timestamp$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER update_timestamp BEFORE UPDATE ON certificate
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE OR REPLACE TRIGGER update_timestamp BEFORE UPDATE ON student
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE OR REPLACE TRIGGER update_timestamp BEFORE UPDATE ON course
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE OR REPLACE TRIGGER update_timestamp BEFORE UPDATE ON template
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
