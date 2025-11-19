import type { FC } from "react";
import { Link } from "react-router-dom";
import { ROUTES } from "../../Routes";
import { Button, Col, Container, Row } from "react-bootstrap";

export const HomePage: FC = () => {
  return (
    <Container>
      <Row>
        <Col md={6}>
          <h1>Классификатор АГ</h1>
          <p>
            Это справочник стадий артериальной гипертензии. Используйте поиск и фильтры по
            систолическому и диастолическому давлению, чтобы найти нужную стадию. Изображения берутся
            из MinIO, при их отсутствии показывается иконка по умолчанию.
          </p>
          <Link to={ROUTES.ALBUMS}>
            <Button variant="primary">Перейти к стадиям</Button>
          </Link>
        </Col>
      </Row>
    </Container>
  );
};