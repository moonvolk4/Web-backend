import "./AlbumPage.css";
import type { FC } from "react";
import { useEffect, useState } from "react";
import { BreadCrumbs } from "../components/BreadCrumbs";
import { ROUTES, ROUTE_LABELS } from "../../Routes";
import { useParams } from "react-router-dom";
import type { ITunesMusic } from "../modules/itunesApi";
import { getAlbumById, getStageByIdRaw } from "../modules/itunesApi";
import { Col, Row, Spinner, Image } from "react-bootstrap";
import { SONGS_MOCK } from "../modules/mock";
import defaultimage from "../assets/header_icon.png";

export const AlbumPage: FC = () => {
  const [pageData, setPageDdata] = useState<ITunesMusic>();
  const [raw, setRaw] = useState<any>(null);

  const { id } = useParams(); // ид страницы, пример: "/albums/12"

  useEffect(() => {
    if (!id) return;
    Promise.all([getAlbumById(id), getStageByIdRaw(id)])
      .then(([response, rawStage]) => {
        setPageDdata(response.results[0]);
        setRaw(rawStage);
      })
      .catch(
        () =>
          setPageDdata(
            SONGS_MOCK.results.find(
              (album) => String(album.collectionId) == id
            )
          ) /* В случае ошибки используем мок данные, фильтруем по ид */
      );
  }, [id]);


  return (
    <div>
      <BreadCrumbs
        crumbs={[
          { label: ROUTE_LABELS.ALBUMS, path: ROUTES.ALBUMS },
          { label: pageData?.collectionCensoredName || "Услуга" },
        ]}
      />
      {pageData ? ( // проверка на наличие данных, иначе загрузка
        <div className="container">
          <Row>
            <Col md={6}>
              <p>
                Название: <strong>{pageData.collectionCensoredName}</strong>
              </p>
              {raw?.Description && (
                <p>
                  Описание: <strong>{raw.Description}</strong>
                </p>
              )}
              {raw?.Pressure && (
                <p>
                  Давление: <strong>{raw.Pressure}</strong>
                </p>
              )}
              {raw?.RiskName && (
                <p>
                  Риск: <strong>{raw.RiskName}</strong>
                </p>
              )}
            </Col>
            <Col md={6}>
              <Image
                src={pageData.artworkUrl100 || defaultimage} // дефолтное изображение, если нет artworkUrl100
                alt="Картинка"
                width={100}
              />
            </Col>
          </Row>
        </div>
      ) : (
        <div className="album_page_loader_block">{/* загрузка */}
          <Spinner animation="border" />
        </div>
      )}
    </div>
  );
};