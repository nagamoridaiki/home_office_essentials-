// 自分の情報（GET /me などの応答）。未登録の項目は null
export type Customer = {
  email: string;
  family_name: string | null;
  given_name: string | null;
  family_name_kana: string | null;
  given_name_kana: string | null;
  phone_number: string | null;
};

// API のエラー応答の error の中身
export type ApiError = {
  code: string;
  message: string;
  details: { field: string; code: string; message: string }[];
};
