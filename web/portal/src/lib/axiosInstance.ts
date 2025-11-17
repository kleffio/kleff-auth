import axios, {
  type AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios";

import { extractMessage } from "@/shared/errors/extractMessage";
import {
  GenericHttpException,
  HttpException,
  InvalidInputException,
  NotFoundException,
} from "@/shared/errors/HttpExceptions";

axios.defaults.withCredentials = false;

declare module "axios" {
  export interface AxiosRequestConfig {
    useV2?: boolean;
  }
}

function mapAxiosError(error: AxiosError): HttpException | Error {
  const statusCode = error.response?.status ?? 0;
  const data = error.response?.data;
  const message = extractMessage(
    data,
    error.message || "Unexpected error"
  );

  switch (statusCode) {
    case 404:
      return new NotFoundException(message);
    case 422:
      return new InvalidInputException(message);
    default:
      return new GenericHttpException(statusCode || 500, message);
  }
}

const createAxiosInstance = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: "/api/v1",
    withCredentials: false,
  });

  instance.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      config.baseURL = config.useV2 ? "/api/v2" : "/api/v1";
      return config;
    },
    (error) => Promise.reject(error)
  );

  instance.interceptors.response.use(
    (response) => response,
    (error: AxiosError | Error) => {
      const axiosError = error as AxiosError;

      if (axiosError.request && !axiosError.response) {
        return Promise.reject(
          new GenericHttpException(
            503,
            "Kleff backend is unreachable. Please try again later."
          )
        );
      }

      if (axiosError.response) {
        return Promise.reject(mapAxiosError(axiosError));
      }

      return Promise.reject(
        new GenericHttpException(500, axiosError.message || "Unknown error")
      );
    }
  );

  return instance;
};

const axiosInstance = createAxiosInstance();
export default axiosInstance;
