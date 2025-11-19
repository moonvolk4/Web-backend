import type { FC } from 'react';
import { useEffect, useState } from 'react';
import "./ITunesPage.css";
import { Col, Row, Spinner, Form, Button } from "react-bootstrap";
import type { ITunesMusic } from "../modules/itunesApi";
import { getStages } from "../modules/itunesApi";
import InputField  from "../components/InputField";
import { BreadCrumbs } from "../components/BreadCrumbs";
import { ROUTES, ROUTE_LABELS } from "../../Routes";
import { MusicCard } from "../components/MusicCard";
import { useNavigate } from "react-router-dom";
import { SONGS_MOCK } from "../modules/mock";

const ITunesPage: FC = () => {
  const [searchValue, setSearchValue] = useState("");
  const [loading, setLoading] = useState(false);
  const [music, setMusic] = useState<ITunesMusic[]>([]);
  const [sysFrom, setSysFrom] = useState<string>("");
  const [sysTo, setSysTo] = useState<string>("");
  const [diaFrom, setDiaFrom] = useState<string>("");
  const [diaTo, setDiaTo] = useState<string>("");

  const navigate = useNavigate();

  const handleSearch = () => {
    setLoading(true);
    getStages({
      query: searchValue,
      sys_from: sysFrom ? Number(sysFrom) : undefined,
      sys_to: sysTo ? Number(sysTo) : undefined,
      dia_from: diaFrom ? Number(diaFrom) : undefined,
      dia_to: diaTo ? Number(diaTo) : undefined,
    })
      .then((response) => {
        setMusic(response.results);
        setLoading(false);
      })
      .catch(() => { // В случае ошибки используем mock данные, фильтруем по имени
        setMusic(
          SONGS_MOCK.results.filter((item) =>
            item.collectionCensoredName
              .toLocaleLowerCase()
              .startsWith(searchValue.toLocaleLowerCase())
          )
        );
        setLoading(false);
      });
  };

  // Auto-load initial services list on mount
  useEffect(() => {
    handleSearch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

    function handleCardClick(collectionId: number): void {
        navigate(`${ROUTES.ALBUMS}/${collectionId}`);
    }

  return (
    <div className="container-fluid">
      <BreadCrumbs crumbs={[{ label: ROUTE_LABELS.ALBUMS }]} />
      
      <InputField
        value={searchValue}
        setValue={(value) => setSearchValue(value)}
        loading={loading}
        onSubmit={handleSearch}
        placeholder="Название стадии"
        buttonTitle="Искать"
      />

      <Form className="mb-3">
        <Row className="g-2 align-items-end">
          <Col xs={12} sm={6} md={3}>
            <Form.Group controlId="sysFrom">
              <Form.Label>Систолическое от</Form.Label>
              <Form.Control type="number" min="0" value={sysFrom} onChange={(e) => setSysFrom(e.target.value)} placeholder="напр. 120" />
            </Form.Group>
          </Col>
          <Col xs={12} sm={6} md={3}>
            <Form.Group controlId="sysTo">
              <Form.Label>Систолическое до</Form.Label>
              <Form.Control type="number" min="0" value={sysTo} onChange={(e) => setSysTo(e.target.value)} placeholder="напр. 140" />
            </Form.Group>
          </Col>
          <Col xs={12} sm={6} md={3}>
            <Form.Group controlId="diaFrom">
              <Form.Label>Диастолическое от</Form.Label>
              <Form.Control type="number" min="0" value={diaFrom} onChange={(e) => setDiaFrom(e.target.value)} placeholder="напр. 80" />
            </Form.Group>
          </Col>
          <Col xs={12} sm={6} md={3}>
            <Form.Group controlId="diaTo">
              <Form.Label>Диастолическое до</Form.Label>
              <Form.Control type="number" min="0" value={diaTo} onChange={(e) => setDiaTo(e.target.value)} placeholder="напр. 90" />
            </Form.Group>
          </Col>
          <Col xs={12} md={2}>
            <Button variant="secondary" onClick={handleSearch} disabled={loading}>
              Применить фильтры
            </Button>
          </Col>
        </Row>
      </Form>

      {loading && ( // здесь можно было использовать тернарный оператор, но это усложняет читаемость
        <div className="loadingBg">
          <Spinner animation="border" />
        </div>
      )}
      {!loading &&
        (!music.length /* Проверка на существование данных */ ? (
          <div>
            <h1>К сожалению, пока ничего не найдено :(</h1>
          </div>
        ) : (
          <Row xs={1} sm={2} md={3} lg={4} xl={4} xxl={5} className="g-4">
            {music.map((item, index) => (
              <Col key={index}>
                <MusicCard
                  imageClickHandler={() => handleCardClick(item.collectionId)}
                  {...item}
                />
              </Col>
            ))}
          </Row>
        ))}
    </div>
  );
};

export default ITunesPage;