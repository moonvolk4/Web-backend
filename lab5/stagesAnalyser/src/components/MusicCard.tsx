import type { FC } from 'react'
import { Button, Card } from "react-bootstrap";
import { Link } from "react-router-dom";
import "./MusicCard.css";
import defaultimage from "../assets/header_icon.png";

interface ICardProps {
  artworkUrl100: string;
  artistName: string;
  collectionCensoredName: string;
  trackViewUrl: string;
  imageClickHandler: () => void;
  pressure?: string;
  riskName?: string;
  code?: string;
  collectionId?: number;
}

export const MusicCard: FC<ICardProps> = ({
  artworkUrl100,
  artistName,
  collectionCensoredName,
  trackViewUrl,
  imageClickHandler,
  pressure,
  riskName,
  code,
  collectionId,
}) => {
  const riskText = riskName || artistName || '';
  const computedLink = trackViewUrl || (typeof collectionId === 'number' ? `/albums/${collectionId}` : '');

  return (
    <Card className="card">
      <Card.Img
        className="cardImage"
        variant="top"
        src={artworkUrl100 || defaultimage}
        height={100}
        width={100}
        onClick={imageClickHandler}
      />
      <Card.Body>
        <div className="textStyle">
          <Card.Title>{collectionCensoredName}</Card.Title>
        </div>
        {pressure && (
          <div className="textStyle">
            <Card.Text>Давление: {pressure}</Card.Text>
          </div>
        )}
        {riskText && (
          <div className="textStyle">
            <Card.Text>Риск: {riskText}</Card.Text>
          </div>
        )}
        {code && (
          <div className="textStyle">
            <Card.Text>Код: {code}</Card.Text>
          </div>
        )}
        {computedLink && !computedLink.endsWith('undefined') ? (
          <Link to={computedLink} className="btn btn-primary cardButton">
            Открыть
          </Link>
        ) : (
          <Button className="cardButton" variant="secondary" disabled>
            Недоступно
          </Button>
        )}
      </Card.Body>
    </Card>
  );
};