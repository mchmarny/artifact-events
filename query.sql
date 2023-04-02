-- list distinct images
SELECT DISTINCT image FROM `cloudy-demos.artifact.vul` ORDER BY 1

-- list versions a given image
SELECT DISTINCT digest
FROM `cloudy-demos.artifact.vul`
WHERE image = 'https://us-west1-docker.pkg.dev/cloudy-demos/events/test38'