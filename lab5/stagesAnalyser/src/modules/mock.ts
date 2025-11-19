import type { ITunesResult } from "./itunesApi";

export const SONGS_MOCK: ITunesResult = {
  resultCount: 3, 
  results: [
    {
      wrapperType: "track",
      artistName: "Нулевой",
      collectionCensoredName: "Оптимальное",
      trackViewUrl: "",
      artworkUrl100: "",
      collectionId: 1,
      pressure: "120 / 80",
      riskName: "Нулевой",
      code: "I10-1",
    },
    {
      wrapperType: "track",
      artistName: "Минимальный",
      collectionCensoredName: "Нормальное",
      trackViewUrl: "",
      artworkUrl100: "",
      collectionId: 2,
      pressure: "120 - 129 / 80 - 84",
      riskName: "Минимальный",
      code: "I10-1",
    },
    {
      wrapperType: "track",
      artistName: "Незначительный",
      collectionCensoredName: "Высокое",
      trackViewUrl: "",
      artworkUrl100: "",
      collectionId: 3,
      pressure: "130 - 139 / 85 - 89",
      riskName: "Незначительный",
      code: "I10-1",
    },
  ],
};
